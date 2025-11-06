package repository

import (
    "context"

    dbmodel "github.com/Dysar/url-crawler/backend/internal/models/db"
    "gorm.io/gorm"
)

type ResultRepository interface {
    Create(ctx context.Context, res dbmodel.CrawlResult) (*dbmodel.CrawlResult, error)
    GetByURLID(ctx context.Context, urlID int64) (*dbmodel.CrawlResult, error)
}

type resultRepository struct{ db *gorm.DB }

func NewResultRepository(db *gorm.DB) ResultRepository { return &resultRepository{db: db} }

func (r *resultRepository) Create(ctx context.Context, res dbmodel.CrawlResult) (*dbmodel.CrawlResult, error) {
    // Explicitly list columns
    err := r.db.WithContext(ctx).Model(&dbmodel.CrawlResult{}).Select(
        "job_id", "url_id", "html_version", "title", "headings_h1", "headings_h2", "headings_h3",
        "headings_h4", "headings_h5", "headings_h6", "internal_links_count", "external_links_count",
        "inaccessible_links_count", "has_login_form",
    ).Create(&res).Error
    if err != nil {
        return nil, err
    }
    var out dbmodel.CrawlResult
    if err := r.db.WithContext(ctx).Select(
        "id", "job_id", "url_id", "html_version", "title", "headings_h1", "headings_h2", "headings_h3",
        "headings_h4", "headings_h5", "headings_h6", "internal_links_count", "external_links_count",
        "inaccessible_links_count", "has_login_form", "created_at",
    ).First(&out, res.ID).Error; err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *resultRepository) GetByURLID(ctx context.Context, urlID int64) (*dbmodel.CrawlResult, error) {
    var out dbmodel.CrawlResult
    if err := r.db.WithContext(ctx).Select(
        "id", "job_id", "url_id", "html_version", "title", "headings_h1", "headings_h2", "headings_h3",
        "headings_h4", "headings_h5", "headings_h6", "internal_links_count", "external_links_count",
        "inaccessible_links_count", "has_login_form", "created_at",
    ).Where("url_id = ?", urlID).Order("created_at DESC").First(&out).Error; err != nil {
        return nil, err
    }
    return &out, nil
}

