package models

import "time"

type CrawlJobStatus string

const (
	JobQueued    CrawlJobStatus = "queued"
	JobRunning   CrawlJobStatus = "running"
	JobCompleted CrawlJobStatus = "completed"
	JobFailed    CrawlJobStatus = "failed"
	JobStopped   CrawlJobStatus = "stopped" // stopped by the user
)

type CrawlJob struct {
	ID          int64          `db:"id"`
	URLID       int64          `db:"url_id"`
	Status      CrawlJobStatus `db:"status"`
	StartedAt   *time.Time     `db:"started_at"`
	CompletedAt *time.Time     `db:"completed_at"`
	Error       *string        `db:"error_message"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

type URL struct {
	ID        int64     `db:"id"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CrawlResult struct {
	ID                     int64     `db:"id"`
	JobID                  int64     `db:"job_id"`
	URLID                  int64     `db:"url_id"`
	HTMLVersion            *string   `db:"html_version"`
	Title                  *string   `db:"title"`
	HeadingsH1             int       `db:"headings_h1"`
	HeadingsH2             int       `db:"headings_h2"`
	HeadingsH3             int       `db:"headings_h3"`
	HeadingsH4             int       `db:"headings_h4"`
	HeadingsH5             int       `db:"headings_h5"`
	HeadingsH6             int       `db:"headings_h6"`
	InternalLinksCount     int       `db:"internal_links_count"`
	ExternalLinksCount     int       `db:"external_links_count"`
	InaccessibleLinksCount int       `db:"inaccessible_links_count"`
	HasLoginForm           bool      `db:"has_login_form"`
	CreatedAt              time.Time `db:"created_at"`
}
