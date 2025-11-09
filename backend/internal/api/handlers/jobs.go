package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/service"
)

type JobHandlers struct {
	svc *service.JobService
}

func NewJobHandlers(svc *service.JobService) *JobHandlers {
	return &JobHandlers{svc: svc}
}

type startJobsRequest struct {
	URLIDs []int64 `json:"url_ids" binding:"required"`
}

func (h *JobHandlers) Start(c *gin.Context) {
	var req startJobsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.URLIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url_ids required"})
		return
	}
	started := make([]models.JobStartResponse, 0, len(req.URLIDs))
	for _, id := range req.URLIDs {
		urlRec, err := h.svc.GetURLByID(c, id)
		if err != nil || urlRec == nil {
			continue
		}
		jobID, err := h.svc.StartForURL(c, id, urlRec.URL)
		if err == nil {
			started = append(started, models.JobStartResponse{URLID: id, JobID: jobID})
		}
	}
	c.JSON(http.StatusOK, started)
}

func (h *JobHandlers) Status(c *gin.Context) {
	idParam := c.Param("id")
	var id int64
	_, _ = fmt.Sscan(idParam, &id)
	status, err := h.svc.GetJobStatus(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, status)
}

type stopJobsRequest struct {
	URLIDs []int64 `json:"url_ids" binding:"required"`
}

func (h *JobHandlers) Stop(c *gin.Context) {
	var req stopJobsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.URLIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url_ids required"})
		return
	}

	stopped, err := h.svc.StopJobs(c, req.URLIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.JobsStoppedResponse(stopped))
}
