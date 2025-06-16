package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"job-scheduler/internal/scheduler"
	"job-scheduler/pkg/database"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db        *database.Connection
	scheduler *scheduler.Scheduler
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.Connection, scheduler *scheduler.Scheduler) *HealthHandler {
	return &HealthHandler{
		db:        db,
		scheduler: scheduler,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]interface{} `json:"services"`
}

// HealthCheck handles GET /api/v1/health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  make(map[string]interface{}),
	}

	// Check database health
	dbStatus := h.checkDatabaseHealth()
	response.Services["database"] = dbStatus

	// Check scheduler health
	schedulerStatus := h.checkSchedulerHealth()
	response.Services["scheduler"] = schedulerStatus

	// Determine overall status
	if dbStatus["status"] != "healthy" || schedulerStatus["status"] != "healthy" {
		response.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// checkDatabaseHealth checks the database connection health
func (h *HealthHandler) checkDatabaseHealth() map[string]interface{} {
	status := map[string]interface{}{
		"status": "healthy",
	}

	start := time.Now()
	err := h.db.HealthCheck()
	duration := time.Since(start)

	status["response_time_ms"] = duration.Milliseconds()

	if err != nil {
		status["status"] = "unhealthy"
		status["error"] = err.Error()
		logrus.WithError(err).Error("Database health check failed")
	}

	return status
}

// checkSchedulerHealth checks the scheduler health
func (h *HealthHandler) checkSchedulerHealth() map[string]interface{} {
	status := map[string]interface{}{
		"status":         "healthy",
		"is_running":     h.scheduler.IsRunning(),
		"scheduled_jobs": h.scheduler.GetScheduledJobsCount(),
	}

	if !h.scheduler.IsRunning() {
		status["status"] = "unhealthy"
		status["error"] = "Scheduler is not running"
	}

	return status
}

// RegisterRoutes registers health-related routes
func (h *HealthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/health", h.HealthCheck)
}
