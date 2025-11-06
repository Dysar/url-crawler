package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	dbmodel "github.com/Dysar/url-crawler/backend/internal/models/db"
)

type ResultRepository interface {
	Create(ctx context.Context, res dbmodel.CrawlResult) (*dbmodel.CrawlResult, error)
	GetByURLID(ctx context.Context, urlID int64) (*dbmodel.CrawlResult, error)
}

type resultRepository struct {
	db *sqlx.DB
}

func NewResultRepository(db *sqlx.DB) ResultRepository {
	return &resultRepository{db: db}
}

// Create inserts a new crawl result with explicit column selection
// Uses prepared statement for optimal performance
func (r *resultRepository) Create(ctx context.Context, res dbmodel.CrawlResult) (*dbmodel.CrawlResult, error) {
	query := `INSERT INTO crawl_results (
		job_id, url_id, html_version, title, 
		headings_h1, headings_h2, headings_h3, headings_h4, headings_h5, headings_h6,
		internal_links_count, external_links_count, inaccessible_links_count, has_login_form
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(ctx, query,
		res.JobID, res.URLID, res.HTMLVersion, res.Title,
		res.HeadingsH1, res.HeadingsH2, res.HeadingsH3, res.HeadingsH4, res.HeadingsH5, res.HeadingsH6,
		res.InternalLinksCount, res.ExternalLinksCount, res.InaccessibleLinksCount, res.HasLoginForm,
	)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// Fetch the created record with explicit column selection
	var out dbmodel.CrawlResult
	query = `SELECT id, job_id, url_id, html_version, title,
	         headings_h1, headings_h2, headings_h3, headings_h4, headings_h5, headings_h6,
	         internal_links_count, external_links_count, inaccessible_links_count, has_login_form, created_at
	         FROM crawl_results WHERE id = ?`
	if err := r.db.GetContext(ctx, &out, query, id); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetByURLID fetches the most recent result for a URL
// Uses ORDER BY and LIMIT for efficiency
func (r *resultRepository) GetByURLID(ctx context.Context, urlID int64) (*dbmodel.CrawlResult, error) {
	var out dbmodel.CrawlResult
	query := `SELECT id, job_id, url_id, html_version, title,
	          headings_h1, headings_h2, headings_h3, headings_h4, headings_h5, headings_h6,
	          internal_links_count, external_links_count, inaccessible_links_count, has_login_form, created_at
	          FROM crawl_results 
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
