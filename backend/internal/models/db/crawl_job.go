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
    ID          int64          `gorm:"column:id;primaryKey;autoIncrement"`
    URLID       int64          `gorm:"column:url_id;not null"`
    Status      CrawlJobStatus `gorm:"column:status;type:enum('queued','running','completed','failed');default:'queued'"`
    StartedAt   *time.Time     `gorm:"column:started_at"`
    CompletedAt *time.Time     `gorm:"column:completed_at"`
    Error       *string        `gorm:"column:error_message"`
    CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (CrawlJob) TableName() string { return "crawl_jobs" }


