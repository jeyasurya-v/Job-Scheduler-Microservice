package repositories

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"job-scheduler/internal/models"
)

// JobRepository defines the interface for job data operations
type JobRepository interface {
	Create(job *models.Job) error
	GetByID(id uuid.UUID) (*models.Job, error)
	GetAll(page, limit int) ([]models.Job, int64, error)
	Update(job *models.Job) error
	Delete(id uuid.UUID) error
	GetActiveJobs() ([]models.Job, error)
	GetByJobType(jobType models.JobType) ([]models.Job, error)
}

// jobRepository implements JobRepository interface
type jobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{
		db: db,
	}
}

// Create creates a new job in the database
func (r *jobRepository) Create(job *models.Job) error {
	if err := r.db.Create(job).Error; err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	return nil
}

// GetByID retrieves a job by its ID
func (r *jobRepository) GetByID(id uuid.UUID) (*models.Job, error) {
	var job models.Job
	err := r.db.Where("id = ?", id).First(&job).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("job with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get job by ID: %w", err)
	}
	return &job, nil
}

// GetAll retrieves all jobs with pagination
func (r *jobRepository) GetAll(page, limit int) ([]models.Job, int64, error) {
	var jobs []models.Job
	var totalCount int64

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	if err := r.db.Model(&models.Job{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Get jobs with pagination, ordered by created_at desc
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs: %w", err)
	}

	return jobs, totalCount, nil
}

// Update updates an existing job
func (r *jobRepository) Update(job *models.Job) error {
	// Use Select to update all fields including zero values
	err := r.db.Model(job).Select("*").Where("id = ?", job.ID).Updates(job).Error
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Check if any rows were affected
	if r.db.RowsAffected == 0 {
		return fmt.Errorf("job with ID %s not found", job.ID)
	}

	return nil
}

// Delete deletes a job by its ID
func (r *jobRepository) Delete(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.Job{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete job: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("job with ID %s not found", id)
	}

	return nil
}

// GetActiveJobs retrieves all active jobs
func (r *jobRepository) GetActiveJobs() ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.Where("is_active = ?", true).Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active jobs: %w", err)
	}
	return jobs, nil
}

// GetByJobType retrieves jobs by their type
func (r *jobRepository) GetByJobType(jobType models.JobType) ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.Where("job_type = ?", jobType).Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by type: %w", err)
	}
	return jobs, nil
}
