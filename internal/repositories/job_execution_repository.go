package repositories

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"job-scheduler/internal/models"
)

// JobExecutionRepository defines the interface for job execution data operations
type JobExecutionRepository interface {
	Create(execution *models.JobExecution) error
	GetByID(id uuid.UUID) (*models.JobExecution, error)
	GetByJobID(jobID uuid.UUID, page, limit int) ([]models.JobExecution, int64, error)
	Update(execution *models.JobExecution) error
	Delete(id uuid.UUID) error
	GetRunningExecutions() ([]models.JobExecution, error)
	GetExecutionStats(jobID uuid.UUID) (*models.JobExecutionStats, error)
	GetRecentExecutions(limit int) ([]models.JobExecution, error)
}

// jobExecutionRepository implements JobExecutionRepository interface
type jobExecutionRepository struct {
	db *gorm.DB
}

// NewJobExecutionRepository creates a new job execution repository
func NewJobExecutionRepository(db *gorm.DB) JobExecutionRepository {
	return &jobExecutionRepository{
		db: db,
	}
}

// Create creates a new job execution in the database
func (r *jobExecutionRepository) Create(execution *models.JobExecution) error {
	if err := r.db.Create(execution).Error; err != nil {
		return fmt.Errorf("failed to create job execution: %w", err)
	}
	return nil
}

// GetByID retrieves a job execution by its ID
func (r *jobExecutionRepository) GetByID(id uuid.UUID) (*models.JobExecution, error) {
	var execution models.JobExecution
	err := r.db.Preload("Job").Where("id = ?", id).First(&execution).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("job execution with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get job execution by ID: %w", err)
	}
	return &execution, nil
}

// GetByJobID retrieves job executions for a specific job with pagination
func (r *jobExecutionRepository) GetByJobID(jobID uuid.UUID, page, limit int) ([]models.JobExecution, int64, error) {
	var executions []models.JobExecution
	var totalCount int64

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count for the specific job
	if err := r.db.Model(&models.JobExecution{}).Where("job_id = ?", jobID).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count job executions: %w", err)
	}

	// Get executions with pagination, ordered by started_at desc
	err := r.db.Where("job_id = ?", jobID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get job executions: %w", err)
	}

	return executions, totalCount, nil
}

// Update updates an existing job execution
func (r *jobExecutionRepository) Update(execution *models.JobExecution) error {
	err := r.db.Model(execution).Select("*").Where("id = ?", execution.ID).Updates(execution).Error
	if err != nil {
		return fmt.Errorf("failed to update job execution: %w", err)
	}

	if r.db.RowsAffected == 0 {
		return fmt.Errorf("job execution with ID %s not found", execution.ID)
	}

	return nil
}

// Delete deletes a job execution by its ID
func (r *jobExecutionRepository) Delete(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.JobExecution{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete job execution: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("job execution with ID %s not found", id)
	}

	return nil
}

// GetRunningExecutions retrieves all currently running job executions
func (r *jobExecutionRepository) GetRunningExecutions() ([]models.JobExecution, error) {
	var executions []models.JobExecution
	err := r.db.Preload("Job").
		Where("status = ?", models.ExecutionStatusRunning).
		Find(&executions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get running executions: %w", err)
	}
	return executions, nil
}

// GetExecutionStats calculates statistics for job executions of a specific job
func (r *jobExecutionRepository) GetExecutionStats(jobID uuid.UUID) (*models.JobExecutionStats, error) {
	var stats models.JobExecutionStats

	// Get total executions count
	if err := r.db.Model(&models.JobExecution{}).
		Where("job_id = ?", jobID).
		Count(&stats.TotalExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total executions: %w", err)
	}

	// Get successful executions count
	if err := r.db.Model(&models.JobExecution{}).
		Where("job_id = ? AND status = ?", jobID, models.ExecutionStatusCompleted).
		Count(&stats.SuccessfulExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to count successful executions: %w", err)
	}

	// Get failed executions count
	if err := r.db.Model(&models.JobExecution{}).
		Where("job_id = ? AND status = ?", jobID, models.ExecutionStatusFailed).
		Count(&stats.FailedExecutions).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed executions: %w", err)
	}

	// Calculate success rate
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions) * 100
	}

	// Get average execution time for completed jobs
	var avgDuration *float64
	err := r.db.Model(&models.JobExecution{}).
		Select("AVG(execution_duration)").
		Where("job_id = ? AND status = ? AND execution_duration IS NOT NULL", jobID, models.ExecutionStatusCompleted).
		Scan(&avgDuration).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average execution time: %w", err)
	}

	if avgDuration != nil {
		avgDurationInt := int64(*avgDuration)
		stats.AverageExecutionTime = &avgDurationInt
	}

	return &stats, nil
}

// GetRecentExecutions retrieves the most recent job executions across all jobs
func (r *jobExecutionRepository) GetRecentExecutions(limit int) ([]models.JobExecution, error) {
	var executions []models.JobExecution
	err := r.db.Preload("Job").
		Order("started_at DESC").
		Limit(limit).
		Find(&executions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}
	return executions, nil
}
