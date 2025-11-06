package service

import (
	"context"

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

func NewJobService(j repository.JobRepository, r repository.ResultRepository, u repository.URLRepository, c *crawler.Crawler) *JobService {
	return &JobService{jobs: j, results: r, urls: u, craw: c}
}

func (s *JobService) StartForURL(ctx context.Context, urlID int64, url string) (int64, error) {
	job, err := s.jobs.Enqueue(ctx, urlID)
	if err != nil {
		return 0, err
	}
	go s.process(context.Background(), job.ID, urlID, url)
	return job.ID, nil
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
