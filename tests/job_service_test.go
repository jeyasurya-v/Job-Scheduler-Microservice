package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"job-scheduler/internal/models"
	"job-scheduler/internal/services"
)

// MockJobRepository is a mock implementation of JobRepository
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Create(job *models.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepository) GetByID(id uuid.UUID) (*models.Job, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetAll(page, limit int) ([]models.Job, int64, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]models.Job), args.Get(1).(int64), args.Error(2)
}

func (m *MockJobRepository) Update(job *models.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockJobRepository) GetActiveJobs() ([]models.Job, error) {
	args := m.Called()
	return args.Get(0).([]models.Job), args.Error(1)
}

func (m *MockJobRepository) GetByJobType(jobType models.JobType) ([]models.Job, error) {
	args := m.Called(jobType)
	return args.Get(0).([]models.Job), args.Error(1)
}

func TestJobService_CreateJob(t *testing.T) {
	// Setup
	mockRepo := new(MockJobRepository)
	jobService := services.NewJobService(mockRepo)

	// Test data
	req := &models.CreateJobRequest{
		Name:        "Test Job",
		Description: "A test job",
		Schedule:    "0 9 * * *", // Every day at 9 AM
		JobType:     models.JobTypeEmailNotification,
		Config: models.JobConfig{
			"recipient": "test@example.com",
		},
	}

	// Mock expectations
	mockRepo.On("Create", mock.AnythingOfType("*models.Job")).Return(nil)

	// Execute
	job, err := jobService.CreateJob(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, req.Name, job.Name)
	assert.Equal(t, req.Description, job.Description)
	assert.Equal(t, req.Schedule, job.Schedule)
	assert.Equal(t, req.JobType, job.JobType)
	assert.True(t, job.IsActive)
	assert.NotEqual(t, uuid.Nil, job.ID)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestJobService_CreateJob_InvalidCronSchedule(t *testing.T) {
	// Setup
	mockRepo := new(MockJobRepository)
	jobService := services.NewJobService(mockRepo)

	// Test data with invalid cron schedule
	req := &models.CreateJobRequest{
		Name:        "Test Job",
		Description: "A test job",
		Schedule:    "invalid cron", // Invalid cron expression
		JobType:     models.JobTypeEmailNotification,
	}

	// Execute
	job, err := jobService.CreateJob(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "invalid cron schedule")

	// Verify no repository calls were made
	mockRepo.AssertNotCalled(t, "Create")
}

func TestJobService_CreateJob_InvalidJobType(t *testing.T) {
	// Setup
	mockRepo := new(MockJobRepository)
	jobService := services.NewJobService(mockRepo)

	// Test data with invalid job type
	req := &models.CreateJobRequest{
		Name:        "Test Job",
		Description: "A test job",
		Schedule:    "0 9 * * *",
		JobType:     models.JobType("invalid_type"), // Invalid job type
	}

	// Execute
	job, err := jobService.CreateJob(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "invalid job type")

	// Verify no repository calls were made
	mockRepo.AssertNotCalled(t, "Create")
}

func TestJobService_ValidateCronSchedule(t *testing.T) {
	// Setup
	mockRepo := new(MockJobRepository)
	jobService := services.NewJobService(mockRepo)

	// Test cases
	testCases := []struct {
		name     string
		schedule string
		isValid  bool
	}{
		{"Valid - Every minute", "* * * * *", true},
		{"Valid - Every day at 9 AM", "0 9 * * *", true},
		{"Valid - Every Monday at midnight", "0 0 * * 1", true},
		{"Valid - Every 5 minutes", "*/5 * * * *", true},
		{"Invalid - Too many fields", "* * * * * *", false},
		{"Invalid - Too few fields", "* * *", false},
		{"Invalid - Invalid minute", "60 * * * *", false},
		{"Invalid - Invalid hour", "0 25 * * *", false},
		{"Invalid - Invalid day of month", "0 0 32 * *", false},
		{"Invalid - Invalid month", "0 0 1 13 *", false},
		{"Invalid - Invalid day of week", "0 0 * * 8", false},
		{"Invalid - Random text", "not a cron", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := jobService.ValidateCronSchedule(tc.schedule)
			if tc.isValid {
				assert.NoError(t, err, "Expected schedule '%s' to be valid", tc.schedule)
			} else {
				assert.Error(t, err, "Expected schedule '%s' to be invalid", tc.schedule)
			}
		})
	}
}

func TestJobService_GetAllJobs(t *testing.T) {
	// Setup
	mockRepo := new(MockJobRepository)
	jobService := services.NewJobService(mockRepo)

	// Test data
	expectedJobs := []models.Job{
		{
			ID:       uuid.New(),
			Name:     "Job 1",
			JobType:  models.JobTypeEmailNotification,
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Name:     "Job 2",
			JobType:  models.JobTypeDataProcessing,
			IsActive: true,
		},
	}
	expectedCount := int64(2)

	// Mock expectations
	mockRepo.On("GetAll", 1, 10).Return(expectedJobs, expectedCount, nil)

	// Execute
	response, err := jobService.GetAllJobs(1, 10)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedJobs, response.Jobs)
	assert.Equal(t, expectedCount, response.TotalCount)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 1, response.TotalPages)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestJobService_GetAllJobs_PaginationDefaults(t *testing.T) {
	// Setup
	mockRepo := new(MockJobRepository)
	jobService := services.NewJobService(mockRepo)

	// Mock expectations with default pagination
	mockRepo.On("GetAll", 1, 10).Return([]models.Job{}, int64(0), nil)

	// Execute with invalid pagination parameters
	response, err := jobService.GetAllJobs(0, -5) // Invalid page and limit

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.Page)  // Should default to 1
	assert.Equal(t, 10, response.Limit) // Should default to 10

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
