package service

import (
	"context"
	"errors"

	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
)

type ResultService struct {
	repo repository.ResultRepository
}

func NewResultService(repo repository.ResultRepository) (*ResultService, error) {
	if repo == nil {
		return nil, errors.New("ResultRepository must not be nil")
	}
	return &ResultService{repo: repo}, nil
}

func (s *ResultService) GetResultByURLID(ctx context.Context, urlID int64) (*models.ResultResponse, error) {
	res, err := s.repo.GetByURLID(ctx, urlID)
	if err != nil {
		return nil, err
	}
	return &models.ResultResponse{
		ID:                     res.ID,
		URLID:                  res.URLID,
		HTMLVersion:            res.HTMLVersion,
		Title:                  res.Title,
		HeadingsH1:             res.HeadingsH1,
		HeadingsH2:             res.HeadingsH2,
		HeadingsH3:             res.HeadingsH3,
		HeadingsH4:             res.HeadingsH4,
		HeadingsH5:             res.HeadingsH5,
		HeadingsH6:             res.HeadingsH6,
		InternalLinksCount:     res.InternalLinksCount,
		ExternalLinksCount:     res.ExternalLinksCount,
		InaccessibleLinksCount: res.InaccessibleLinksCount,
		HasLoginForm:           res.HasLoginForm,
	}, nil
}

