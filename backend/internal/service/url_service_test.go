package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Dysar/url-crawler/backend/internal/mocks"
	models "github.com/Dysar/url-crawler/backend/internal/models"
)

func TestURLService_CreateURL_HappyPath(t *testing.T) {
	ctx := context.Background()
	url := "https://example.com"
	expectedID := int64(123)
	now := time.Now()

	mockRepo := new(mocks.URLRepository)
	mockRepo.On("Create", ctx, url).Return(&models.URL{
		ID:        expectedID,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	svc, err := NewURLService(mockRepo)
	assert.NoError(t, err)

	resp, err := svc.CreateURL(ctx, url)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedID, resp.ID)
	assert.Equal(t, url, resp.URL)
	mockRepo.AssertExpectations(t)
}

func TestURLService_ListURLs_HappyPath(t *testing.T) {
	ctx := context.Background()
	page := 1
	limit := 20
	sortBy := "created_at"
	order := "desc"
	total := int64(2)
	now := time.Now()

	mockRepo := new(mocks.URLRepository)
	mockRepo.On("List", ctx, page, limit, sortBy, order).Return([]models.URL{
		{
			ID:        1,
			URL:       "https://example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        2,
			URL:       "https://test.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, total, nil)

	svc, err := NewURLService(mockRepo)
	assert.NoError(t, err, "NewURLService should not return error")

	resp, err := svc.ListURLs(ctx, page, limit, sortBy, order)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, total, resp.Total)
	assert.Equal(t, page, resp.Page)
	assert.Equal(t, limit, resp.Limit)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, int64(1), resp.Data[0].ID)
	assert.Equal(t, "https://example.com", resp.Data[0].URL)
	assert.Equal(t, int64(2), resp.Data[1].ID)
	assert.Equal(t, "https://test.com", resp.Data[1].URL)
	mockRepo.AssertExpectations(t)
}
