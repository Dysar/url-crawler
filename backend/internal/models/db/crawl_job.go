package db

import "time"

type CrawlJobStatus string

const (
    JobQueued    CrawlJobStatus = "queued"
    JobRunning   CrawlJobStatus = "running"
    JobCompleted CrawlJobStatus = "completed"
    JobFailed    CrawlJobStatus = "failed"
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


