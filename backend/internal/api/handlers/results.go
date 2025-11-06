package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
)

type ResultHandlers struct {
	results repository.ResultRepository
}

func NewResultHandlers(results repository.ResultRepository) *ResultHandlers {
	return &ResultHandlers{results: results}
}

func (h *ResultHandlers) GetByURLID(c *gin.Context) {
	idParam := c.Param("id")
	urlID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url_id"})
		return
	}
	res, err := h.results.GetByURLID(c, urlID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, models.ResultResponse{
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
	})
}
