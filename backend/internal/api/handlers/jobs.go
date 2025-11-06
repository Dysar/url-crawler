package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	models "github.com/Dysar/url-crawler/backend/internal/models"
	"github.com/Dysar/url-crawler/backend/internal/repository"
	"github.com/Dysar/url-crawler/backend/internal/service"
)

type JobHandlers struct {
	urls repository.URLRepository
	jobs repository.JobRepository
	svc  *service.JobService
}

func NewJobHandlers(urls repository.URLRepository, jobs repository.JobRepository, svc *service.JobService) *JobHandlers {
	return &JobHandlers{urls: urls, jobs: jobs, svc: svc}
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
		urlRec, err := h.urls.GetByID(c, id)
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
	// minimal status by id
	// parse id
	idParam := c.Param("id")
	var id int64
	_, _ = fmt.Sscan(idParam, &id)
	job, err := h.jobs.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, models.JobStatusResponse{
		ID:     job.ID,
		Status: job.Status,
		Error:  job.Error,
	})
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

	stopped := make([]models.JobsStoppedItem, 0)
	stopMsg := "Stopped by user"

	for _, urlID := range req.URLIDs {
		// Find the most recent job for this URL
		// For simplicity, we'll need to add a method to get job by URL ID
		// Or we can query by URL ID and get the latest job
		// Let's add a helper method to job repository
		job, err := h.jobs.GetByURLID(c, urlID)
		if err != nil {
			continue
		}

		// Only stop queued or running jobs
		if job.Status == models.JobQueued || job.Status == models.JobRunning {
			if err := h.jobs.UpdateStatus(c, job.ID, models.JobFailed, &stopMsg); err == nil {
				stopped = append(stopped, models.JobsStoppedItem{URLID: urlID, JobID: job.ID})
			}
		}
	}

	c.JSON(http.StatusOK, models.JobsStoppedResponse(stopped))
}
