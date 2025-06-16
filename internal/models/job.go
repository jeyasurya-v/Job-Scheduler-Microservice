package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobType represents the type of job to be executed
type JobType string

const (
	JobTypeEmailNotification JobType = "email_notification"
	JobTypeDataProcessing    JobType = "data_processing"
	JobTypeReportGeneration  JobType = "report_generation"
	JobTypeHealthCheck       JobType = "health_check"
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusActive   JobStatus = "active"
	JobStatusInactive JobStatus = "inactive"
)

// JobConfig holds configuration data for different job types
// This is stored as JSONB in PostgreSQL for flexibility
type JobConfig map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (jc JobConfig) Value() (driver.Value, error) {
	if jc == nil {
		return nil, nil
	}
	return json.Marshal(jc)
}

// Scan implements the sql.Scanner interface for database retrieval
func (jc *JobConfig) Scan(value interface{}) error {
	if value == nil {
		*jc = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JobConfig", value)
	}

	return json.Unmarshal(bytes, jc)
}

// Job represents a scheduled job in the system
type Job struct {
	// Primary key - using UUID for better scalability
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Basic job information
	Name        string `json:"name" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Description string `json:"description" gorm:"type:text"`

	// Scheduling information
	Schedule string `json:"schedule" gorm:"not null;size:100" validate:"required,cron"`

	// Job type and configuration
	JobType JobType   `json:"job_type" gorm:"not null;size:50" validate:"required,oneof=email_notification data_processing report_generation health_check"`
	Config  JobConfig `json:"config" gorm:"type:jsonb"`

	// Status and metadata
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Executions []JobExecution `json:"executions,omitempty" gorm:"foreignKey:JobID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate is a GORM hook that runs before creating a job
func (j *Job) BeforeCreate(tx *gorm.DB) error {
	// Generate UUID if not provided
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Job model
func (Job) TableName() string {
	return "jobs"
}

// IsValidJobType checks if the job type is valid
func IsValidJobType(jobType string) bool {
	switch JobType(jobType) {
	case JobTypeEmailNotification, JobTypeDataProcessing, JobTypeReportGeneration, JobTypeHealthCheck:
		return true
	default:
		return false
	}
}

// GetDefaultConfig returns default configuration for each job type
func GetDefaultConfig(jobType JobType) JobConfig {
	switch jobType {
	case JobTypeEmailNotification:
		return JobConfig{
			"recipient": "user@example.com",
			"subject":   "Scheduled Notification",
			"body":      "This is a scheduled email notification.",
		}
	case JobTypeDataProcessing:
		return JobConfig{
			"processing_time_seconds": 5,
			"data_size":              "1MB",
			"operation":              "transform",
		}
	case JobTypeReportGeneration:
		return JobConfig{
			"report_type": "daily_summary",
			"format":      "txt",
			"include_charts": false,
		}
	case JobTypeHealthCheck:
		return JobConfig{
			"url":             "https://httpbin.org/status/200",
			"timeout_seconds": 30,
			"expected_status": 200,
		}
	default:
		return JobConfig{}
	}
}

// CreateJobRequest represents the request payload for creating a job
type CreateJobRequest struct {
	Name        string    `json:"name" validate:"required,min=1,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	Schedule    string    `json:"schedule" validate:"required"`
	JobType     JobType   `json:"job_type" validate:"required"`
	Config      JobConfig `json:"config"`
	IsActive    *bool     `json:"is_active"` // Pointer to distinguish between false and nil
}

// UpdateJobRequest represents the request payload for updating a job
type UpdateJobRequest struct {
	Name        *string    `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description" validate:"omitempty,max=1000"`
	Schedule    *string    `json:"schedule" validate:"omitempty"`
	JobType     *JobType   `json:"job_type" validate:"omitempty"`
	Config      *JobConfig `json:"config"`
	IsActive    *bool      `json:"is_active"`
}

// JobListResponse represents the response for listing jobs with pagination
type JobListResponse struct {
	Jobs       []Job `json:"jobs"`
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}
