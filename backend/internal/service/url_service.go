package service

import (
	"context"
	"errors"

	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) (*URLService, error) {
	if repo == nil {
		return nil, errors.New("URLRepository must not be nil")
	}
	return &URLService{repo: repo}, nil
}

func (s *URLService) CreateURL(ctx context.Context, url string) (*models.URLResponse, error) {
	rec, err := s.repo.Create(ctx, url)
	if err != nil {
		return nil, err
	}
	return &models.URLResponse{ID: rec.ID, URL: rec.URL}, nil
}

func (s *URLService) ListURLs(ctx context.Context, page int, limit int, sortBy string, order string) (*models.URLListResponse, error) {
	rows, total, err := s.repo.List(ctx, page, limit, sortBy, order)
	if err != nil {
		return nil, err
	}
	resp := make([]models.URLResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, models.URLResponse{ID: r.ID, URL: r.URL})
	}
	return &models.URLListResponse{
		Data:  resp,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *URLService) GetURLByID(ctx context.Context, id int64) (*models.URL, error) {
	return s.repo.GetByID(ctx, id)
}

