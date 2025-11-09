package service

import (
	"context"
	"errors"

	"github.com/Dysar/url-crawler/backend/internal/crawler"
	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
)

type JobService struct {
	jobs    repository.JobRepository
	results repository.ResultRepository
	urls    repository.URLRepository
	craw    *crawler.Crawler
}

func NewJobService(j repository.JobRepository, r repository.ResultRepository, u repository.URLRepository, c *crawler.Crawler) (*JobService, error) {
	if j == nil || r == nil || u == nil {
		return nil, errors.New("all deps for job service must be not nil")
	}
	return &JobService{jobs: j, results: r, urls: u, craw: c}, nil
}

func (s *JobService) StartForURL(ctx context.Context, urlID int64, url string) (int64, error) {
	job, err := s.jobs.Enqueue(ctx, urlID)
	if err != nil {
		return 0, err
	}
	go s.process(context.Background(), job.ID, urlID, url)
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
			if err := s.jobs.UpdateStatus(ctx, job.ID, models.JobFailed, &stopMsg); err == nil {
				stopped = append(stopped, models.JobsStoppedItem{URLID: urlID, JobID: job.ID})
			}
		}
	}

	return stopped, nil
}

func (s *JobService) GetURLByID(ctx context.Context, urlID int64) (*models.URL, error) {
	return s.urls.GetByID(ctx, urlID)
}

func (s *JobService) process(ctx context.Context, jobID int64, urlID int64, target string) {
	_ = s.jobs.UpdateStatus(ctx, jobID, models.JobRunning, nil)
	res, err := s.craw.Crawl(ctx, target)
	if err != nil {
		msg := err.Error()
		_ = s.jobs.UpdateStatus(ctx, jobID, models.JobFailed, &msg)
		return
	}
	// Persist
	var htmlVer, title *string
	htmlVer = res.HTMLVersion
	title = res.Title
	_, _ = s.results.Create(ctx, models.CrawlResult{
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
	_ = s.jobs.UpdateStatus(ctx, jobID, models.JobCompleted, nil)
}
