package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	dbmodel "github.com/Dysar/url-crawler/backend/internal/models/db"
)

type JobRepository interface {
	Enqueue(ctx context.Context, urlID int64) (*dbmodel.CrawlJob, error)
	UpdateStatus(ctx context.Context, id int64, status dbmodel.CrawlJobStatus, errMsg *string) error
	GetByID(ctx context.Context, id int64) (*dbmodel.CrawlJob, error)
	GetByURLID(ctx context.Context, urlID int64) (*dbmodel.CrawlJob, error)
}

type jobRepository struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) JobRepository {
	return &jobRepository{db: db}
}

// Enqueue creates a new crawl job with status 'queued'
// Uses prepared statement for optimal performance
func (r *jobRepository) Enqueue(ctx context.Context, urlID int64) (*dbmodel.CrawlJob, error) {
	query := `INSERT INTO crawl_jobs (url_id, status) VALUES (?, ?)`
	result, err := r.db.ExecContext(ctx, query, urlID, dbmodel.JobQueued)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Fetch the created record with explicit column selection
	var out dbmodel.CrawlJob
	query = `SELECT id, url_id, status, started_at, completed_at, error_message, created_at, updated_at 
	         FROM crawl_jobs WHERE id = ?`
	if err := r.db.GetContext(ctx, &out, query, id); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateStatus updates job status with optimized queries based on status transition
// Uses prepared statements and only updates necessary fields
func (r *jobRepository) UpdateStatus(ctx context.Context, id int64, status dbmodel.CrawlJobStatus, errMsg *string) error {
	now := time.Now()

	var query string
	var args []interface{}

	switch status {
	case dbmodel.JobRunning:
		// Set started_at when job starts running
		query = `UPDATE crawl_jobs SET status = ?, started_at = ?, updated_at = ? WHERE id = ?`
		args = []interface{}{status, now, now, id}
	case dbmodel.JobCompleted, dbmodel.JobFailed:
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
func (r *jobRepository) GetByID(ctx context.Context, id int64) (*dbmodel.CrawlJob, error) {
	var out dbmodel.CrawlJob
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
func (r *jobRepository) GetByURLID(ctx context.Context, urlID int64) (*dbmodel.CrawlJob, error) {
	var out dbmodel.CrawlJob
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
