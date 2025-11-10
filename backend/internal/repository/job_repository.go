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

	now := time.Now()
	return &models.CrawlJob{
		ID:        id,
		URLID:     urlID,
		Status:    models.JobQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateStatus updates job status with optimized queries based on status transition
// Uses prepared statements and only updates necessary fields
// Uses atomic updates with WHERE clauses to prevent race conditions in concurrent scenarios
/*
	Example:
	Goroutine 1 (Worker): Tries to update "queued" → "running"
	Goroutine 2 (User):   Tries to update "queued" → "stopped"
                      (happens first)
	Result: Goroutine 1's update fails (sql.ErrNoRows) because status is no longer "queued"
        This is expected and handled gracefully
*/
func (r *jobRepository) UpdateStatus(ctx context.Context, id int64, status models.CrawlJobStatus, errMsg *string) error {
	now := time.Now()

	// Build SET clause
	setClause := "status = ?, updated_at = ?"
	setArgs := []any{status, now}

	switch status {
	case models.JobRunning:
		setClause += ", started_at = ?"
		setArgs = append(setArgs, now)
	case models.JobCompleted, models.JobFailed, models.JobStopped:
		setClause += ", completed_at = ?"
		setArgs = append(setArgs, now)
		if errMsg != nil {
			setClause += ", error_message = ?"
			setArgs = append(setArgs, *errMsg)
		}
	}

	// Build WHERE clause with status precondition
	var whereClause string
	var whereArgs []any
	switch status {
	case models.JobRunning:
		whereClause = "id = ? AND status = ?"
		whereArgs = []any{id, models.JobQueued}
	case models.JobCompleted, models.JobFailed:
		whereClause = "id = ? AND status = ?"
		whereArgs = []any{id, models.JobRunning}
	case models.JobStopped:
		whereClause = "id = ? AND status IN (?, ?)"
		whereArgs = []any{id, models.JobQueued, models.JobRunning}
	default:
		whereClause = "id = ?"
		whereArgs = []any{id}
	}

	query := "UPDATE crawl_jobs SET " + setClause + " WHERE " + whereClause
	args := append(setArgs, whereArgs...)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
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
