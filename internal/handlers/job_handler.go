package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"job-scheduler/internal/models"
	"job-scheduler/internal/services"
)

// JobHandler handles HTTP requests for job operations
type JobHandler struct {
	jobService services.JobService
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobService services.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

// CreateJob handles POST /api/v1/jobs
func (h *JobHandler) CreateJob(c *gin.Context) {
	var req models.CreateJobRequest

	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind create job request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job name is required",
		})
		return
	}

	if req.Schedule == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job schedule is required",
		})
		return
	}

	if req.JobType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job type is required",
		})
		return
	}

	// Create job
	job, err := h.jobService.CreateJob(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create job")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create job",
			"details": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"job_name": job.Name,
	}).Info("Job created via API")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Job created successfully",
		"job":     job,
	})
}

// GetJob handles GET /api/v1/jobs/{id}
func (h *JobHandler) GetJob(c *gin.Context) {
	// Parse job ID from URL parameter
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
		})
		return
	}

	// Get job
	job, err := h.jobService.GetJobByID(jobID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get job")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Job not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job": job,
	})
}

// GetJobs handles GET /api/v1/jobs
func (h *JobHandler) GetJobs(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get jobs
	response, err := h.jobService.GetAllJobs(page, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get jobs")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve jobs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateJob handles PUT /api/v1/jobs/{id}
func (h *JobHandler) UpdateJob(c *gin.Context) {
	// Parse job ID from URL parameter
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
		})
		return
	}

	var req models.UpdateJobRequest

	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind update job request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update job
	job, err := h.jobService.UpdateJob(jobID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update job")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update job",
			"details": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"job_name": job.Name,
	}).Info("Job updated via API")

	c.JSON(http.StatusOK, gin.H{
		"message": "Job updated successfully",
		"job":     job,
	})
}

// DeleteJob handles DELETE /api/v1/jobs/{id}
func (h *JobHandler) DeleteJob(c *gin.Context) {
	// Parse job ID from URL parameter
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
		})
		return
	}

	// Delete job
	if err := h.jobService.DeleteJob(jobID); err != nil {
		logrus.WithError(err).Error("Failed to delete job")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to delete job",
			"details": err.Error(),
		})
		return
	}

	logrus.WithField("job_id", jobID).Info("Job deleted via API")

	c.JSON(http.StatusOK, gin.H{
		"message": "Job deleted successfully",
	})
}

// RegisterRoutes registers all job-related routes
func (h *JobHandler) RegisterRoutes(router *gin.RouterGroup) {
	jobs := router.Group("/jobs")
	{
		jobs.POST("", h.CreateJob)
		jobs.GET("", h.GetJobs)
		jobs.GET("/:id", h.GetJob)
		jobs.PUT("/:id", h.UpdateJob)
		jobs.DELETE("/:id", h.DeleteJob)
	}
}
