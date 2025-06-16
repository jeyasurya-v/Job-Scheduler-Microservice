package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"job-scheduler/internal/config"
	"job-scheduler/internal/models"
	"job-scheduler/internal/repositories"
	"job-scheduler/internal/services"
)

// JobExecutor handles the execution of individual jobs
type JobExecutor struct {
	jobExecutionRepo repositories.JobExecutionRepository
	executors        map[models.JobType]services.JobExecutor
	config           *config.Config
	semaphore        chan struct{} // Limits concurrent job executions
	mu               sync.RWMutex
	runningJobs      map[uuid.UUID]*models.JobExecution
}

// NewJobExecutor creates a new job executor
func NewJobExecutor(jobExecutionRepo repositories.JobExecutionRepository, cfg *config.Config) *JobExecutor {
	// Create semaphore to limit concurrent executions
	semaphore := make(chan struct{}, cfg.Scheduler.MaxConcurrentJobs)

	// Initialize job type executors
	executors := map[models.JobType]services.JobExecutor{
		models.JobTypeEmailNotification: &services.EmailNotificationExecutor{},
		models.JobTypeDataProcessing:    &services.DataProcessingExecutor{},
		models.JobTypeReportGeneration:  services.NewReportGenerationExecutor(cfg.Reports.Directory),
		models.JobTypeHealthCheck:       services.NewHealthCheckExecutor(cfg.HealthCheck.Timeout),
	}

	return &JobExecutor{
		jobExecutionRepo: jobExecutionRepo,
		executors:        executors,
		config:           cfg,
		semaphore:        semaphore,
		runningJobs:      make(map[uuid.UUID]*models.JobExecution),
	}
}

// ExecuteJob executes a job with proper error handling and logging
func (e *JobExecutor) ExecuteJob(job *models.Job) error {
	// Acquire semaphore to limit concurrent executions
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	default:
		logrus.WithFields(logrus.Fields{
			"job_id":   job.ID,
			"job_name": job.Name,
		}).Warn("Job execution skipped - maximum concurrent jobs reached")
		return fmt.Errorf("maximum concurrent jobs (%d) reached", e.config.Scheduler.MaxConcurrentJobs)
	}

	// Create job execution record
	execution := &models.JobExecution{
		ID:     uuid.New(),
		JobID:  job.ID,
		Status: models.ExecutionStatusPending,
	}

	// Save initial execution record
	if err := e.jobExecutionRepo.Create(execution); err != nil {
		logrus.WithFields(logrus.Fields{
			"job_id": job.ID,
			"error":  err,
		}).Error("Failed to create job execution record")
		return fmt.Errorf("failed to create execution record: %w", err)
	}

	// Track running job
	e.mu.Lock()
	e.runningJobs[execution.ID] = execution
	e.mu.Unlock()

	// Clean up tracking when done
	defer func() {
		e.mu.Lock()
		delete(e.runningJobs, execution.ID)
		e.mu.Unlock()
	}()

	// Execute job with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Execute in goroutine to handle timeout
	errChan := make(chan error, 1)
	go func() {
		errChan <- e.executeJobWithContext(ctx, job, execution)
	}()

	// Wait for completion or timeout
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		execution.MarkAsFailed("Job execution timed out")
		if updateErr := e.jobExecutionRepo.Update(execution); updateErr != nil {
			logrus.WithFields(logrus.Fields{
				"execution_id": execution.ID,
				"error":        updateErr,
			}).Error("Failed to update execution record after timeout")
		}
		return fmt.Errorf("job execution timed out")
	}
}

// executeJobWithContext executes a job with the given context
func (e *JobExecutor) executeJobWithContext(ctx context.Context, job *models.Job, execution *models.JobExecution) error {
	// Mark execution as running
	execution.MarkAsRunning()
	if err := e.jobExecutionRepo.Update(execution); err != nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"error":        err,
		}).Error("Failed to update execution status to running")
	}

	logrus.WithFields(logrus.Fields{
		"job_id":       job.ID,
		"job_name":     job.Name,
		"job_type":     job.JobType,
		"execution_id": execution.ID,
	}).Info("Starting job execution")

	// Get executor for job type
	executor, exists := e.executors[job.JobType]
	if !exists {
		err := fmt.Errorf("no executor found for job type: %s", job.JobType)
		execution.MarkAsFailed(err.Error())
		if updateErr := e.jobExecutionRepo.Update(execution); updateErr != nil {
			logrus.WithFields(logrus.Fields{
				"execution_id": execution.ID,
				"error":        updateErr,
			}).Error("Failed to update execution record")
		}
		return err
	}

	// Execute the job
	var executionErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				executionErr = fmt.Errorf("job execution panicked: %v", r)
				logrus.WithFields(logrus.Fields{
					"job_id":       job.ID,
					"execution_id": execution.ID,
					"panic":        r,
				}).Error("Job execution panicked")
			}
		}()

		// Check context before execution
		select {
		case <-ctx.Done():
			executionErr = ctx.Err()
			return
		default:
		}

		// Execute the job
		executionErr = executor.Execute(job)
	}()

	// Update execution status based on result
	if executionErr != nil {
		execution.MarkAsFailed(executionErr.Error())
		logrus.WithFields(logrus.Fields{
			"job_id":       job.ID,
			"job_name":     job.Name,
			"execution_id": execution.ID,
			"error":        executionErr,
		}).Error("Job execution failed")
	} else {
		execution.MarkAsCompleted()
		logrus.WithFields(logrus.Fields{
			"job_id":            job.ID,
			"job_name":          job.Name,
			"execution_id":      execution.ID,
			"execution_duration": execution.GetDurationString(),
		}).Info("Job execution completed successfully")
	}

	// Save final execution status
	if err := e.jobExecutionRepo.Update(execution); err != nil {
		logrus.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"error":        err,
		}).Error("Failed to update final execution status")
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	return executionErr
}

// GetRunningJobs returns a list of currently running job executions
func (e *JobExecutor) GetRunningJobs() []*models.JobExecution {
	e.mu.RLock()
	defer e.mu.RUnlock()

	running := make([]*models.JobExecution, 0, len(e.runningJobs))
	for _, execution := range e.runningJobs {
		running = append(running, execution)
	}

	return running
}

// GetRunningJobsCount returns the number of currently running jobs
func (e *JobExecutor) GetRunningJobsCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.runningJobs)
}

// GetMaxConcurrentJobs returns the maximum number of concurrent jobs allowed
func (e *JobExecutor) GetMaxConcurrentJobs() int {
	return e.config.Scheduler.MaxConcurrentJobs
}
