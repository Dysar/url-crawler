package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Dysar/url-crawler/backend/internal/crawler"
	"github.com/Dysar/url-crawler/backend/internal/mocks"
	models "github.com/Dysar/url-crawler/backend/internal/models"
)

// createTestCrawler creates a crawler with an httptest server for testing
func createTestCrawler(html string) (*crawler.Crawler, *httptest.Server) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	c := crawler.New(crawler.HTTPClient(5 * time.Second))
	return c, ts
}

func TestJobService_StartForURL_HappyPath(t *testing.T) {
	ctx := context.Background()
	urlID := int64(123)
	url := "http://example.com"
	expectedJobID := int64(456)

	mockJobs := new(mocks.JobRepository)
	mockJobs.On("Enqueue", ctx, urlID).Return(&models.CrawlJob{
		ID:     expectedJobID,
		URLID:  urlID,
		Status: models.JobQueued,
	}, nil)
	// StartForURL launches a goroutine that calls process, but we don't test that here
	// The goroutine will try to call UpdateStatus, so we need to allow it (but it may fail)
	// Use Maybe() to allow the call but not require it
	mockJobs.On("UpdateStatus", mock.Anything, expectedJobID, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockResults := new(mocks.ResultRepository)
	mockResults.On("Create", mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	mockURLs := new(mocks.URLRepository)
	realCrawler := crawler.New(crawler.HTTPClient(1 * time.Second))

	svc := NewJobService(mockJobs, mockResults, mockURLs, realCrawler)

	jobID, err := svc.StartForURL(ctx, urlID, url)

	assert.NoError(t, err, "StartForURL should not return error")
	assert.Equal(t, expectedJobID, jobID, "should return the correct job ID")
	// Wait a bit for goroutine to potentially complete
	time.Sleep(200 * time.Millisecond)
	mockJobs.AssertExpectations(t)
}

func TestJobService_Process_HappyPath(t *testing.T) {
	ctx := context.Background()
	jobID := int64(456)
	urlID := int64(123)

	htmlVer := "HTML5"
	title := "Test Page"

	// Create test server with HTML that will produce expected results
	html := `<!doctype html>
<html>
<head><title>Test Page</title></head>
<body>
<h1>Heading 1</h1><h1>Heading 1</h1>
<h2>Heading 2</h2><h2>Heading 2</h2><h2>Heading 2</h2>
<h3>Heading 3</h3>
<a href="/internal1">Internal</a>
<a href="/internal2">Internal</a>
<a href="https://external.com">External</a>
<input type="password" name="pwd" />
</body>
</html>`
	realCrawler, ts := createTestCrawler(html)
	defer ts.Close()

	mockJobs := new(mocks.JobRepository)
	mockJobs.On("UpdateStatus", mock.Anything, jobID, models.JobRunning, (*string)(nil)).Return(nil)
	mockJobs.On("UpdateStatus", mock.Anything, jobID, models.JobCompleted, (*string)(nil)).Return(nil)

	mockResults := new(mocks.ResultRepository)
	mockResults.On("Create", mock.Anything, mock.MatchedBy(func(res models.CrawlResult) bool {
		return res.JobID == jobID &&
			res.URLID == urlID &&
			res.HTMLVersion != nil && *res.HTMLVersion == htmlVer &&
			res.Title != nil && *res.Title == title &&
			res.HeadingsH1 == 2 &&
			res.HeadingsH2 == 3 &&
			res.HeadingsH3 == 1 &&
			res.InternalLinksCount == 2 &&
			res.ExternalLinksCount == 1 &&
			res.HasLoginForm == true
	})).Return(&models.CrawlResult{
		ID:    1,
		JobID: jobID,
		URLID: urlID,
	}, nil)

	mockURLs := new(mocks.URLRepository)

	svc := NewJobService(mockJobs, mockResults, mockURLs, realCrawler)

	// Call process directly (not via goroutine) for deterministic testing
	svc.process(ctx, jobID, urlID, ts.URL)

	// Verify all expectations were met
	mockJobs.AssertExpectations(t)
	mockResults.AssertExpectations(t)
}
