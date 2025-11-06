package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
)

type URLHandlers struct {
	repo repository.URLRepository
}

func NewURLHandlers(repo repository.URLRepository) *URLHandlers {
	return &URLHandlers{repo: repo}
}

func (h *URLHandlers) CreateURL(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	rec, err := h.repo.Create(c, req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"data":    models.URLResponse{ID: rec.ID, URL: rec.URL},
		"error":   nil,
		"message": "Success",
	})
}

func (h *URLHandlers) ListURLs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort_by", "created_at")
	order := c.DefaultQuery("order", "desc")

	rows, total, err := h.repo.List(c, page, limit, sortBy, order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := make([]models.URLResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, models.URLResponse{ID: r.ID, URL: r.URL})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    resp,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"error":   nil,
		"message": "Success",
	})
}
