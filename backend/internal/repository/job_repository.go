package repository

import (
    "context"
    "time"

    dbmodel "github.com/Dysar/url-crawler/backend/internal/models/db"
    "gorm.io/gorm"
)

type JobRepository interface {
    Enqueue(ctx context.Context, urlID int64) (*dbmodel.CrawlJob, error)
    UpdateStatus(ctx context.Context, id int64, status dbmodel.CrawlJobStatus, errMsg *string) error
    GetByID(ctx context.Context, id int64) (*dbmodel.CrawlJob, error)
}

type jobRepository struct{ db *gorm.DB }

func NewJobRepository(db *gorm.DB) JobRepository { return &jobRepository{db: db} }

func (r *jobRepository) Enqueue(ctx context.Context, urlID int64) (*dbmodel.CrawlJob, error) {
    rec := &dbmodel.CrawlJob{URLID: urlID, Status: dbmodel.JobQueued}
    if err := r.db.WithContext(ctx).Select("url_id", "status").Create(rec).Error; err != nil {
        return nil, err
    }
    var out dbmodel.CrawlJob
    if err := r.db.WithContext(ctx).
        Select("id", "url_id", "status", "started_at", "completed_at", "error_message", "created_at", "updated_at").
        First(&out, rec.ID).Error; err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id int64, status dbmodel.CrawlJobStatus, errMsg *string) error {
    updates := map[string]interface{}{
        "status": status,
        "updated_at": time.Now(),
    }
    switch status {
    case dbmodel.JobRunning:
        now := time.Now()
        updates["started_at"] = &now
    case dbmodel.JobCompleted, dbmodel.JobFailed:
        now := time.Now()
        updates["completed_at"] = &now
        if errMsg != nil {
            updates["error_message"] = errMsg
        }
    }
    return r.db.WithContext(ctx).Model(&dbmodel.CrawlJob{}).Where("id = ?", id).Updates(updates).Error
}

func (r *jobRepository) GetByID(ctx context.Context, id int64) (*dbmodel.CrawlJob, error) {
    var out dbmodel.CrawlJob
    if err := r.db.WithContext(ctx).
        Select("id", "url_id", "status", "started_at", "completed_at", "error_message", "created_at", "updated_at").
        First(&out, id).Error; err != nil {
        return nil, err
    }
    return &out, nil
}


