package repository

//go:generate mockery --name=JobRepository --output=../mocks --outpkg=mocks

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	models "github.com/Dysar/url-crawler/backend/internal/models"
)

type JobRepository interface {
	Enqueue(ctx context.Context, urlID int64) (*models.CrawlJob, error)
	UpdateStatus(ctx context.Context, id int64, status models.CrawlJobStatus, errMsg *string) error
	GetByID(ctx context.Context, id int64) (*models.CrawlJob, error)
	GetByURLID(ctx context.Context, urlID int64) (*models.CrawlJob, error)
}

type jobRepository struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) JobRepository {
	return &jobRepository{db: db}
}

// Enqueue creates a new crawl job with status 'queued'
// Uses prepared statement for optimal performance
func (r *jobRepository) Enqueue(ctx context.Context, urlID int64) (*models.CrawlJob, error) {
	query := `INSERT INTO crawl_jobs (url_id, status) VALUES (?, ?)`
	result, err := r.db.ExecContext(ctx, query, urlID, models.JobQueued)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Fetch the created record with explicit column selection
	var out models.CrawlJob
	query = `SELECT id, url_id, status, started_at, completed_at, error_message, created_at, updated_at 
	         FROM crawl_jobs WHERE id = ?`
	if err := r.db.GetContext(ctx, &out, query, id); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateStatus updates job status with optimized queries based on status transition
// Uses prepared statements and only updates necessary fields
func (r *jobRepository) UpdateStatus(ctx context.Context, id int64, status models.CrawlJobStatus, errMsg *string) error {
	now := time.Now()

	var query string
	var args []interface{}

	switch status {
	case models.JobRunning:
		// Set started_at when job starts running
		query = `UPDATE crawl_jobs SET status = ?, started_at = ?, updated_at = ? WHERE id = ?`
		args = []interface{}{status, now, now, id}
	case models.JobCompleted, models.JobFailed:
		// Set completed_at when job finishes
		if errMsg != nil {
			query = `UPDATE crawl_jobs SET status = ?, completed_at = ?, error_message = ?, updated_at = ? WHERE id = ?`
			args = []interface{}{status, now, *errMsg, now, id}
		} else {
			query = `UPDATE crawl_jobs SET status = ?, completed_at = ?, updated_at = ? WHERE id = ?`
			args = []interface{}{status, now, now, id}
		}
	default:
		// For queued status, just update status and updated_at
		query = `UPDATE crawl_jobs SET status = ?, updated_at = ? WHERE id = ?`
		args = []interface{}{status, now, id}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetByID fetches a job by ID using prepared statement
func (r *jobRepository) GetByID(ctx context.Context, id int64) (*models.CrawlJob, error) {
	var out models.CrawlJob
	query := `SELECT id, url_id, status, started_at, completed_at, error_message, created_at, updated_at 
	          FROM crawl_jobs WHERE id = ?`
	if err := r.db.GetContext(ctx, &out, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &out, nil
}

// GetByURLID fetches the most recent job for a URL
// Uses ORDER BY and LIMIT for efficiency
func (r *jobRepository) GetByURLID(ctx context.Context, urlID int64) (*models.CrawlJob, error) {
	var out models.CrawlJob
	query := `SELECT id, url_id, status, started_at, completed_at, error_message, created_at, updated_at 
	          FROM crawl_jobs 
	          WHERE url_id = ? 
	          ORDER BY created_at DESC 
	          LIMIT 1`
	if err := r.db.GetContext(ctx, &out, query, urlID); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &out, nil
}
