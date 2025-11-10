package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/Dysar/url-crawler/backend/internal/crawler"
	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
)

type jobTask struct {
	jobID int64
	urlID int64
	url   string
}

type JobService struct {
	jobs    repository.JobRepository
	results repository.ResultRepository
	urls    repository.URLRepository
	craw    *crawler.Crawler

	// Worker pool for parallel job processing
	jobQueue chan jobTask
	workers  int
	wg       sync.WaitGroup
	once     sync.Once
	stop     chan struct{}
}

func NewJobService(j repository.JobRepository, r repository.ResultRepository, u repository.URLRepository, c *crawler.Crawler) (*JobService, error) {
	if j == nil || r == nil || u == nil {
		return nil, errors.New("all deps for job service must be not nil")
	}

	// Default to 10 concurrent workers, can be made configurable
	workers := 10
	queueSize := 100

	svc := &JobService{
		jobs:     j,
		results:  r,
		urls:     u,
		craw:     c,
		jobQueue: make(chan jobTask, queueSize),
		workers:  workers,
		stop:     make(chan struct{}),
	}

	// Start worker pool
	svc.startWorkers()

	return svc, nil
}

// startWorkers starts the worker pool goroutines
func (s *JobService) startWorkers() {
	s.once.Do(func() {
		for i := 0; i < s.workers; i++ {
			s.wg.Add(1)
			go s.worker(i)
		}
	})
}

// worker processes jobs from the queue in parallel
func (s *JobService) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.jobQueue:
			if err := s.process(context.Background(), task.jobID, task.urlID, task.url); err != nil {
				// Log error and attempt to persist to DB
				logrus.WithError(err).WithFields(logrus.Fields{
					"worker_id": id,
					"job_id":    task.jobID,
					"url_id":    task.urlID,
					"url":       task.url,
				}).Error("Failed to process job")

				// Try to persist error to DB (best effort)
				msg := err.Error()
				if dbErr := s.jobs.UpdateStatus(context.Background(), task.jobID, models.JobFailed, &msg); dbErr != nil {
					logrus.WithError(dbErr).WithField("job_id", task.jobID).Error("Failed to persist job error to database")
				}
			}
		case <-s.stop:
			return
		}
	}
}

// Shutdown gracefully shuts down the worker pool
func (s *JobService) Shutdown() {
	close(s.stop)
	s.wg.Wait()
}

func (s *JobService) StartForURL(ctx context.Context, urlID int64, url string) (int64, error) {
	job, err := s.jobs.Enqueue(ctx, urlID)
	if err != nil {
		return 0, err
	}

	// Enqueue job for parallel processing through worker pool
	// This will block if the queue is full, ensuring we respect the concurrency limit
	select {
	case s.jobQueue <- jobTask{jobID: job.ID, urlID: urlID, url: url}:
		// Job queued successfully, will be processed by a worker
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	return job.ID, nil
}

func (s *JobService) GetJobStatus(ctx context.Context, jobID int64) (*models.JobStatusResponse, error) {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return &models.JobStatusResponse{
		ID:     job.ID,
		Status: job.Status,
		Error:  job.Error,
	}, nil
}

func (s *JobService) StopJobs(ctx context.Context, urlIDs []int64) ([]models.JobsStoppedItem, error) {
	stopped := make([]models.JobsStoppedItem, 0)
	stopMsg := "Stopped by user"

	for _, urlID := range urlIDs {
		job, err := s.jobs.GetByURLID(ctx, urlID)
		if err != nil {
			continue
		}

		// Only stop queued or running jobs
		if job.Status == models.JobQueued || job.Status == models.JobRunning {
			if err := s.jobs.UpdateStatus(ctx, job.ID, models.JobStopped, &stopMsg); err == nil {
				stopped = append(stopped, models.JobsStoppedItem{URLID: urlID, JobID: job.ID})
			} else {
				return nil, fmt.Errorf("failed to stop job %d for URL %d: %w", job.ID, urlID, err)
			}
		}
	}

	return stopped, nil
}

func (s *JobService) GetURLByID(ctx context.Context, urlID int64) (*models.URL, error) {
	return s.urls.GetByID(ctx, urlID)
}

// process executes a crawl job and returns an error if any step fails
// Errors are logged and persisted to the database by the caller (worker)
func (s *JobService) process(ctx context.Context, jobID int64, urlID int64, target string) error {
	// Update job status to running
	// If this fails with sql.ErrNoRows, it means the job was already stopped/updated by another goroutine
	if err := s.jobs.UpdateStatus(ctx, jobID, models.JobRunning, nil); err != nil {
		if err == sql.ErrNoRows {
			// Job was likely stopped or updated by another goroutine - this is expected in concurrent scenarios
			return fmt.Errorf("job was already updated by another goroutine (likely stopped)")
		}
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	// Execute crawl
	res, err := s.craw.Crawl(ctx, target)
	if err != nil {
		// Crawl failed - error will be persisted by worker
		return fmt.Errorf("crawl failed: %w", err)
	}

	// Persist results
	var htmlVer, title *string
	htmlVer = res.HTMLVersion
	title = res.Title
	_, err = s.results.Create(ctx, models.CrawlResult{
		JobID:                  jobID,
		URLID:                  urlID,
		HTMLVersion:            htmlVer,
		Title:                  title,
		HeadingsH1:             res.Headings["h1"],
		HeadingsH2:             res.Headings["h2"],
		HeadingsH3:             res.Headings["h3"],
		HeadingsH4:             res.Headings["h4"],
		HeadingsH5:             res.Headings["h5"],
		HeadingsH6:             res.Headings["h6"],
		InternalLinksCount:     res.InternalLinks,
		ExternalLinksCount:     res.ExternalLinks,
		InaccessibleLinksCount: res.InaccessibleLinks,
		HasLoginForm:           res.HasLoginForm,
	})
	if err != nil {
		return fmt.Errorf("failed to persist crawl results: %w", err)
	}

	// Update job status to completed
	// If this fails with sql.ErrNoRows, it means the job was already stopped/updated by another goroutine
	if err := s.jobs.UpdateStatus(ctx, jobID, models.JobCompleted, nil); err != nil {
		if err == sql.ErrNoRows {
			// Job was likely stopped by user while we were processing - this is expected
			return fmt.Errorf("job was stopped while processing")
		}
		return fmt.Errorf("failed to update job status to completed: %w", err)
	}

	return nil
}
