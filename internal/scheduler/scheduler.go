package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"job-scheduler/internal/config"
	"job-scheduler/internal/models"
	"job-scheduler/internal/repositories"
	"job-scheduler/internal/services"
)

// Scheduler manages the execution of scheduled jobs
type Scheduler struct {
	cron                *cron.Cron
	jobService          services.JobService
	jobExecutionRepo    repositories.JobExecutionRepository
	executor            *JobExecutor
	config              *config.Config
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	mu                  sync.RWMutex
	scheduledJobs       map[string]cron.EntryID // job_id -> cron entry id
	isRunning           bool
}

// NewScheduler creates a new job scheduler
func NewScheduler(
	jobService services.JobService,
	jobExecutionRepo repositories.JobExecutionRepository,
	cfg *config.Config,
) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Create cron scheduler with logger
	cronLogger := cron.VerbosePrintfLogger(logrus.StandardLogger())
	c := cron.New(
		cron.WithLogger(cronLogger),
		cron.WithChain(cron.Recover(cronLogger)),
	)

	// Create job executor
	executor := NewJobExecutor(jobExecutionRepo, cfg)

	return &Scheduler{
		cron:             c,
		jobService:       jobService,
		jobExecutionRepo: jobExecutionRepo,
		executor:         executor,
		config:           cfg,
		ctx:              ctx,
		cancel:           cancel,
		scheduledJobs:    make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler and loads all active jobs
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	logrus.Info("Starting job scheduler...")

	// Load and schedule all active jobs
	if err := s.loadActiveJobs(); err != nil {
		return fmt.Errorf("failed to load active jobs: %w", err)
	}

	// Start the cron scheduler
	s.cron.Start()
	s.isRunning = true

	// Start background goroutine to periodically reload jobs
	s.wg.Add(1)
	go s.reloadJobsPeriodically()

	logrus.WithField("scheduled_jobs", len(s.scheduledJobs)).Info("Job scheduler started successfully")
	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping job scheduler...")

	// Cancel context to stop background goroutines
	s.cancel()

	// Stop cron scheduler
	ctx := s.cron.Stop()
	<-ctx.Done() // Wait for running jobs to complete

	// Wait for background goroutines to finish
	s.wg.Wait()

	s.isRunning = false
	logrus.Info("Job scheduler stopped successfully")
	return nil
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(job *models.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !job.IsActive {
		logrus.WithField("job_id", job.ID).Debug("Skipping inactive job")
		return nil
	}

	// Remove existing job if it exists
	if entryID, exists := s.scheduledJobs[job.ID.String()]; exists {
		s.cron.Remove(entryID)
		delete(s.scheduledJobs, job.ID.String())
	}

	// Create job function
	jobFunc := s.createJobFunction(job)

	// Add job to cron scheduler
	entryID, err := s.cron.AddFunc(job.Schedule, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add job to scheduler: %w", err)
	}

	// Store entry ID for later removal
	s.scheduledJobs[job.ID.String()] = entryID

	logrus.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"name":     job.Name,
		"schedule": job.Schedule,
		"entry_id": entryID,
	}).Info("Job added to scheduler")

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.scheduledJobs[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.scheduledJobs, jobID)

		logrus.WithFields(logrus.Fields{
			"job_id":   jobID,
			"entry_id": entryID,
		}).Info("Job removed from scheduler")
	}
}

// GetScheduledJobsCount returns the number of currently scheduled jobs
func (s *Scheduler) GetScheduledJobsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.scheduledJobs)
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// loadActiveJobs loads all active jobs from the database and schedules them
func (s *Scheduler) loadActiveJobs() error {
	jobs, err := s.jobService.GetActiveJobs()
	if err != nil {
		return fmt.Errorf("failed to get active jobs: %w", err)
	}

	logrus.WithField("job_count", len(jobs)).Info("Loading active jobs...")

	for _, job := range jobs {
		if err := s.AddJob(&job); err != nil {
			logrus.WithFields(logrus.Fields{
				"job_id": job.ID,
				"name":   job.Name,
				"error":  err,
			}).Error("Failed to add job to scheduler")
			continue
		}
	}

	return nil
}

// reloadJobsPeriodically periodically reloads jobs from the database
func (s *Scheduler) reloadJobsPeriodically() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Minute) // Reload every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.reloadJobs(); err != nil {
				logrus.WithError(err).Error("Failed to reload jobs")
			}
		}
	}
}

// reloadJobs reloads all active jobs from the database
func (s *Scheduler) reloadJobs() error {
	logrus.Debug("Reloading jobs from database...")

	jobs, err := s.jobService.GetActiveJobs()
	if err != nil {
		return fmt.Errorf("failed to get active jobs: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a map of current jobs for comparison
	currentJobs := make(map[string]*models.Job)
	for _, job := range jobs {
		currentJobs[job.ID.String()] = &job
	}

	// Remove jobs that are no longer active or don't exist
	for jobID, entryID := range s.scheduledJobs {
		if _, exists := currentJobs[jobID]; !exists {
			s.cron.Remove(entryID)
			delete(s.scheduledJobs, jobID)
			logrus.WithField("job_id", jobID).Info("Removed inactive job from scheduler")
		}
	}

	// Add or update jobs
	for _, job := range jobs {
		if job.IsActive {
			// Remove existing entry if it exists
			if entryID, exists := s.scheduledJobs[job.ID.String()]; exists {
				s.cron.Remove(entryID)
				delete(s.scheduledJobs, job.ID.String())
			}

			// Add job with current configuration
			jobFunc := s.createJobFunction(&job)
			entryID, err := s.cron.AddFunc(job.Schedule, jobFunc)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"job_id": job.ID,
					"error":  err,
				}).Error("Failed to add job during reload")
				continue
			}

			s.scheduledJobs[job.ID.String()] = entryID
		}
	}

	logrus.WithField("scheduled_jobs", len(s.scheduledJobs)).Debug("Jobs reloaded successfully")
	return nil
}

// createJobFunction creates a function that executes a specific job
func (s *Scheduler) createJobFunction(job *models.Job) func() {
	return func() {
		// Create a copy of the job to avoid race conditions
		jobCopy := *job

		logrus.WithFields(logrus.Fields{
			"job_id":   jobCopy.ID,
			"name":     jobCopy.Name,
			"job_type": jobCopy.JobType,
		}).Info("Executing scheduled job")

		// Execute the job
		if err := s.executor.ExecuteJob(&jobCopy); err != nil {
			logrus.WithFields(logrus.Fields{
				"job_id": jobCopy.ID,
				"name":   jobCopy.Name,
				"error":  err,
			}).Error("Job execution failed")
		}
	}
}
