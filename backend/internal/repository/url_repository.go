package repository

import (
    "context"

    dbmodel "github.com/Dysar/url-crawler/backend/internal/models/db"
    "gorm.io/gorm"
)

type URLRepository interface {
    Create(ctx context.Context, url string) (*dbmodel.URL, error)
    List(ctx context.Context, page int, limit int, sortBy string, order string) ([]dbmodel.URL, int64, error)
    GetByID(ctx context.Context, id int64) (*dbmodel.URL, error)
}

type urlRepository struct {
    db *gorm.DB
}

func NewURLRepository(db *gorm.DB) URLRepository {
    return &urlRepository{db: db}
}

func (r *urlRepository) Create(ctx context.Context, url string) (*dbmodel.URL, error) {
    rec := &dbmodel.URL{URL: url}
    if err := r.db.WithContext(ctx).Select("url").Create(rec).Error; err != nil {
        return nil, err
    }
    // Load selected columns explicitly to honor no SELECT * rule
    var out dbmodel.URL
    if err := r.db.WithContext(ctx).
        Select("id", "url", "created_at", "updated_at").
        First(&out, rec.ID).Error; err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *urlRepository) List(ctx context.Context, page int, limit int, sortBy string, order string) ([]dbmodel.URL, int64, error) {
    if page < 1 {
        page = 1
    }
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    if sortBy == "" {
        sortBy = "created_at"
    }
    if order != "asc" && order != "desc" {
        order = "desc"
    }

    var total int64
    if err := r.db.WithContext(ctx).Model(&dbmodel.URL{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    var rows []dbmodel.URL
    if err := r.db.WithContext(ctx).
        Select("id", "url", "created_at", "updated_at").
        Order(sortBy + " " + order).
        Offset((page-1)*limit).
        Limit(limit).
        Find(&rows).Error; err != nil {
        return nil, 0, err
    }
    return rows, total, nil
}

func (r *urlRepository) GetByID(ctx context.Context, id int64) (*dbmodel.URL, error) {
    var out dbmodel.URL
    if err := r.db.WithContext(ctx).
        Select("id", "url", "created_at", "updated_at").
        First(&out, id).Error; err != nil {
        return nil, err
    }
    return &out, nil
}

