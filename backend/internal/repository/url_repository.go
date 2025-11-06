package repository

//go:generate mockery --name=URLRepository --output=../mocks --outpkg=mocks

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	models "github.com/Dysar/url-crawler/backend/internal/models"
)

type URLRepository interface {
	Create(ctx context.Context, url string) (*models.URL, error)
	List(ctx context.Context, page int, limit int, sortBy string, order string) ([]models.URL, int64, error)
	GetByID(ctx context.Context, id int64) (*models.URL, error)
}

type urlRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) URLRepository {
	return &urlRepository{db: db}
}

// Create inserts a new URL and returns it with all fields
// Uses prepared statement for optimal performance
func (r *urlRepository) Create(ctx context.Context, url string) (*models.URL, error) {
	query := `INSERT INTO urls (url) VALUES (?)`
	result, err := r.db.ExecContext(ctx, query, url)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Fetch the created record with explicit column selection
	var out models.URL
	query = `SELECT id, url, created_at, updated_at FROM urls WHERE id = ?`
	if err := r.db.GetContext(ctx, &out, query, id); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns paginated URLs with total count
// Uses optimized approach: separate count query (fast with index) + paginated select
// Validates and sanitizes sortBy to prevent SQL injection
func (r *urlRepository) List(ctx context.Context, page int, limit int, sortBy string, order string) ([]models.URL, int64, error) {
	// Validate and sanitize inputs
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Whitelist allowed sort columns to prevent SQL injection
	allowedSorts := map[string]bool{
		"id": true, "url": true, "created_at": true, "updated_at": true,
	}
	if !allowedSorts[sortBy] {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Get total count (optimized with index on created_at)
	var total int64
	countQuery := `SELECT COUNT(*) FROM urls`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	// Fetch paginated results with explicit column selection
	query := `SELECT id, url, created_at, updated_at 
	          FROM urls 
	          ORDER BY ` + sortBy + ` ` + order + `
	          LIMIT ? OFFSET ?`

	var results []models.URL
	if err := r.db.SelectContext(ctx, &results, query, limit, (page-1)*limit); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetByID fetches a URL by ID using prepared statement
func (r *urlRepository) GetByID(ctx context.Context, id int64) (*models.URL, error) {
	var out models.URL
	query := `SELECT id, url, created_at, updated_at FROM urls WHERE id = ?`
	if err := r.db.GetContext(ctx, &out, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &out, nil
}
