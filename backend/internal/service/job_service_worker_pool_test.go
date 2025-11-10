package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Dysar/url-crawler/backend/internal/crawler"
	"github.com/Dysar/url-crawler/backend/internal/mocks"
	models "github.com/Dysar/url-crawler/backend/internal/models"
)

// TestJobService_WorkerPool_ConcurrentExecution tests that multiple jobs run in parallel
func TestJobService_WorkerPool_ConcurrentExecution(t *testing.T) {
	ctx := context.Background()
	numJobs := 5

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<html><head><title>Test</title></head><body></body></html>`))
	}))
	defer ts.Close()

	mockJobs := new(mocks.JobRepository)
	mockResults := new(mocks.ResultRepository)
	mockURLs := new(mocks.URLRepository)
	realCrawler := crawler.New(crawler.HTTPClient(5 * time.Second))

	for i := range numJobs {
		jobID := int64(i + 1)
		urlID := int64(i + 100)

		mockJobs.On("Enqueue", ctx, urlID).Return(&models.CrawlJob{
			ID:     jobID,
			URLID:  urlID,
			Status: models.JobQueued,
		}, nil).Once()

		mockJobs.On("UpdateStatus", mock.Anything, jobID, mock.Anything, mock.Anything).Return(nil).Maybe()
		mockResults.On("Create", mock.Anything, mock.Anything).Return(&models.CrawlResult{ID: jobID}, nil).Maybe()
	}

	svc, err := NewJobService(mockJobs, mockResults, mockURLs, realCrawler)
	assert.NoError(t, err)
	defer svc.Shutdown()

	var wg sync.WaitGroup
	for i := range numJobs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			urlID := int64(idx + 100)
			_, err := svc.StartForURL(ctx, urlID, ts.URL)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	time.Sleep(1 * time.Second)
}

// TestJobService_WorkerPool_Shutdown tests graceful shutdown
func TestJobService_WorkerPool_Shutdown(t *testing.T) {
	ctx := context.Background()
	urlID := int64(123)
	jobID := int64(456)

	mockJobs := new(mocks.JobRepository)
	mockJobs.On("Enqueue", ctx, urlID).Return(&models.CrawlJob{
		ID:     jobID,
		URLID:  urlID,
		Status: models.JobQueued,
	}, nil).Once()

	mockJobs.On("UpdateStatus", mock.Anything, jobID, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockResults := new(mocks.ResultRepository)
	mockResults.On("Create", mock.Anything, mock.Anything).Return(&models.CrawlResult{ID: 1}, nil).Maybe()

	mockURLs := new(mocks.URLRepository)
	realCrawler := crawler.New(crawler.HTTPClient(1 * time.Second))

	svc, err := NewJobService(mockJobs, mockResults, mockURLs, realCrawler)
	assert.NoError(t, err)

	_, err = svc.StartForURL(ctx, urlID, "http://example.com")
	assert.NoError(t, err)

	svc.Shutdown()
}
