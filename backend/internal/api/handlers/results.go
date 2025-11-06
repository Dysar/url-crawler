package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":                      res.ID,
			"url_id":                  res.URLID,
			"html_version":            res.HTMLVersion,
			"title":                   res.Title,
			"headings_h1":             res.HeadingsH1,
			"headings_h2":             res.HeadingsH2,
			"headings_h3":             res.HeadingsH3,
			"headings_h4":             res.HeadingsH4,
			"headings_h5":             res.HeadingsH5,
			"headings_h6":             res.HeadingsH6,
			"internal_links_count":    res.InternalLinksCount,
			"external_links_count":    res.ExternalLinksCount,
			"inaccessible_links_count": res.InaccessibleLinksCount,
			"has_login_form":          res.HasLoginForm,
		},
		"message": "Success",
	})
}

