package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionStatus represents the status of a job execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// JobExecution represents a single execution of a scheduled job
type JobExecution struct {
	// Primary key
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Foreign key to Job
	JobID uuid.UUID `json:"job_id" gorm:"type:uuid;not null;index"`

	// Execution timing
	StartedAt   time.Time  `json:"started_at" gorm:"not null"`
	CompletedAt *time.Time `json:"completed_at"`

	// Execution status and results
	Status       ExecutionStatus `json:"status" gorm:"not null;size:20;default:'pending'"`
	ErrorMessage *string         `json:"error_message" gorm:"type:text"`

	// Performance metrics
	ExecutionDuration *int64 `json:"execution_duration_ms"` // Duration in milliseconds

	// Metadata
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Job Job `json:"job,omitempty" gorm:"foreignKey:JobID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate is a GORM hook that runs before creating a job execution
func (je *JobExecution) BeforeCreate(tx *gorm.DB) error {
	// Generate UUID if not provided
	if je.ID == uuid.Nil {
		je.ID = uuid.New()
	}

	// Set started_at if not provided and status is running
	if je.StartedAt.IsZero() && je.Status == ExecutionStatusRunning {
		je.StartedAt = time.Now().UTC()
	}

	return nil
}

// TableName returns the table name for the JobExecution model
func (JobExecution) TableName() string {
	return "job_executions"
}

// MarkAsRunning updates the execution status to running and sets the start time
func (je *JobExecution) MarkAsRunning() {
	je.Status = ExecutionStatusRunning
	je.StartedAt = time.Now().UTC()
}

// MarkAsCompleted updates the execution status to completed and calculates duration
func (je *JobExecution) MarkAsCompleted() {
	now := time.Now().UTC()
	je.Status = ExecutionStatusCompleted
	je.CompletedAt = &now

	// Calculate execution duration in milliseconds
	if !je.StartedAt.IsZero() {
		duration := now.Sub(je.StartedAt).Milliseconds()
		je.ExecutionDuration = &duration
	}
}

// MarkAsFailed updates the execution status to failed with an error message
func (je *JobExecution) MarkAsFailed(errorMsg string) {
	now := time.Now().UTC()
	je.Status = ExecutionStatusFailed
	je.CompletedAt = &now
	je.ErrorMessage = &errorMsg

	// Calculate execution duration in milliseconds
	if !je.StartedAt.IsZero() {
		duration := now.Sub(je.StartedAt).Milliseconds()
		je.ExecutionDuration = &duration
	}
}

// MarkAsCancelled updates the execution status to cancelled
func (je *JobExecution) MarkAsCancelled() {
	now := time.Now().UTC()
	je.Status = ExecutionStatusCancelled
	je.CompletedAt = &now

	// Calculate execution duration in milliseconds
	if !je.StartedAt.IsZero() {
		duration := now.Sub(je.StartedAt).Milliseconds()
		je.ExecutionDuration = &duration
	}
}

// IsCompleted returns true if the execution has completed (successfully or with failure)
func (je *JobExecution) IsCompleted() bool {
	return je.Status == ExecutionStatusCompleted ||
		je.Status == ExecutionStatusFailed ||
		je.Status == ExecutionStatusCancelled
}

// IsRunning returns true if the execution is currently running
func (je *JobExecution) IsRunning() bool {
	return je.Status == ExecutionStatusRunning
}

// GetDurationString returns a human-readable duration string
func (je *JobExecution) GetDurationString() string {
	if je.ExecutionDuration == nil {
		return "N/A"
	}

	duration := time.Duration(*je.ExecutionDuration) * time.Millisecond
	return duration.String()
}

// JobExecutionListResponse represents the response for listing job executions
type JobExecutionListResponse struct {
	Executions []JobExecution `json:"executions"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// JobExecutionStats represents statistics about job executions
type JobExecutionStats struct {
	TotalExecutions     int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions    int64   `json:"failed_executions"`
	AverageExecutionTime *int64  `json:"average_execution_time_ms"`
	SuccessRate         float64 `json:"success_rate"`
}
