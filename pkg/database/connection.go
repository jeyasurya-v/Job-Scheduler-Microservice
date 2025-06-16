package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"job-scheduler/internal/config"
	"job-scheduler/internal/models"
)

// Connection holds the database connection and configuration
type Connection struct {
	DB     *gorm.DB
	Config *config.Config
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*Connection, error) {
	// Configure GORM logger based on application environment
	var gormLogger logger.Interface
	if cfg.App.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logrus.Info("Successfully connected to database")

	return &Connection{
		DB:     db,
		Config: cfg,
	}, nil
}

// AutoMigrate runs database migrations
func (c *Connection) AutoMigrate() error {
	logrus.Info("Running database migrations...")

	// Run auto-migrations for all models
	err := c.DB.AutoMigrate(
		&models.Job{},
		&models.JobExecution{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	logrus.Info("Database connection closed")
	return nil
}

// HealthCheck performs a database health check
func (c *Connection) HealthCheck() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
