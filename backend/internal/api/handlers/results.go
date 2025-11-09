package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Dysar/url-crawler/backend/internal/service"
)

type ResultHandlers struct {
	svc *service.ResultService
}

func NewResultHandlers(svc *service.ResultService) *ResultHandlers {
	return &ResultHandlers{svc: svc}
}

func (h *ResultHandlers) GetByURLID(c *gin.Context) {
	idParam := c.Param("id")
	urlID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url_id"})
		return
	}
	res, err := h.svc.GetResultByURLID(c, urlID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}
