package services

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"job-scheduler/internal/models"
	"job-scheduler/internal/repositories"
)

// JobService defines the interface for job business logic
type JobService interface {
	CreateJob(req *models.CreateJobRequest) (*models.Job, error)
	GetJobByID(id uuid.UUID) (*models.Job, error)
	GetAllJobs(page, limit int) (*models.JobListResponse, error)
	UpdateJob(id uuid.UUID, req *models.UpdateJobRequest) (*models.Job, error)
	DeleteJob(id uuid.UUID) error
	GetActiveJobs() ([]models.Job, error)
	ValidateCronSchedule(schedule string) error
}

// jobService implements JobService interface
type jobService struct {
	jobRepo repositories.JobRepository
	parser  cron.Parser
}

// NewJobService creates a new job service
func NewJobService(jobRepo repositories.JobRepository) JobService {
	// Create cron parser with standard options
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	return &jobService{
		jobRepo: jobRepo,
		parser:  parser,
	}
}

// CreateJob creates a new job with validation
func (s *jobService) CreateJob(req *models.CreateJobRequest) (*models.Job, error) {
	logrus.WithFields(logrus.Fields{
		"name":     req.Name,
		"job_type": req.JobType,
		"schedule": req.Schedule,
	}).Info("Creating new job")

	// Validate job type
	if !models.IsValidJobType(string(req.JobType)) {
		return nil, fmt.Errorf("invalid job type: %s", req.JobType)
	}

	// Validate cron schedule
	if err := s.ValidateCronSchedule(req.Schedule); err != nil {
		return nil, fmt.Errorf("invalid cron schedule: %w", err)
	}

	// Create job model
	job := &models.Job{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Schedule:    req.Schedule,
		JobType:     req.JobType,
		Config:      req.Config,
		IsActive:    true, // Default to active
	}

	// Override IsActive if provided
	if req.IsActive != nil {
		job.IsActive = *req.IsActive
	}

	// Set default config if not provided
	if job.Config == nil {
		job.Config = models.GetDefaultConfig(req.JobType)
	}

	// Save to database
	if err := s.jobRepo.Create(job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"name":     job.Name,
		"job_type": job.JobType,
	}).Info("Job created successfully")

	return job, nil
}

// GetJobByID retrieves a job by its ID
func (s *jobService) GetJobByID(id uuid.UUID) (*models.Job, error) {
	job, err := s.jobRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return job, nil
}

// GetAllJobs retrieves all jobs with pagination
func (s *jobService) GetAllJobs(page, limit int) (*models.JobListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10 // Default limit
	}

	jobs, totalCount, err := s.jobRepo.GetAll(page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	return &models.JobListResponse{
		Jobs:       jobs,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateJob updates an existing job
func (s *jobService) UpdateJob(id uuid.UUID, req *models.UpdateJobRequest) (*models.Job, error) {
	logrus.WithFields(logrus.Fields{
		"job_id": id,
	}).Info("Updating job")

	// Get existing job
	job, err := s.jobRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job for update: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		job.Name = *req.Name
	}
	if req.Description != nil {
		job.Description = *req.Description
	}
	if req.Schedule != nil {
		// Validate new schedule
		if err := s.ValidateCronSchedule(*req.Schedule); err != nil {
			return nil, fmt.Errorf("invalid cron schedule: %w", err)
		}
		job.Schedule = *req.Schedule
	}
	if req.JobType != nil {
		// Validate new job type
		if !models.IsValidJobType(string(*req.JobType)) {
			return nil, fmt.Errorf("invalid job type: %s", *req.JobType)
		}
		job.JobType = *req.JobType
	}
	if req.Config != nil {
		job.Config = *req.Config
	}
	if req.IsActive != nil {
		job.IsActive = *req.IsActive
	}

	// Save updated job
	if err := s.jobRepo.Update(job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"job_id": job.ID,
		"name":   job.Name,
	}).Info("Job updated successfully")

	return job, nil
}

// DeleteJob deletes a job by its ID
func (s *jobService) DeleteJob(id uuid.UUID) error {
	logrus.WithFields(logrus.Fields{
		"job_id": id,
	}).Info("Deleting job")

	if err := s.jobRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"job_id": id,
	}).Info("Job deleted successfully")

	return nil
}

// GetActiveJobs retrieves all active jobs
func (s *jobService) GetActiveJobs() ([]models.Job, error) {
	jobs, err := s.jobRepo.GetActiveJobs()
	if err != nil {
		return nil, fmt.Errorf("failed to get active jobs: %w", err)
	}
	return jobs, nil
}

// ValidateCronSchedule validates a cron schedule expression
func (s *jobService) ValidateCronSchedule(schedule string) error {
	_, err := s.parser.Parse(schedule)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %w", schedule, err)
	}
	return nil
}
