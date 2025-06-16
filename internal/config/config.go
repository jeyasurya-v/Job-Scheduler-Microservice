package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds all configuration for our application
type Config struct {
	// Database configuration
	Database DatabaseConfig

	// Server configuration
	Server ServerConfig

	// Application configuration
	App AppConfig

	// Scheduler configuration
	Scheduler SchedulerConfig

	// Health check configuration
	HealthCheck HealthCheckConfig

	// Reports configuration
	Reports ReportsConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
}

// AppConfig holds general application configuration
type AppConfig struct {
	Environment string
	LogLevel    string
}

// SchedulerConfig holds scheduler-related configuration
type SchedulerConfig struct {
	Enabled           bool
	MaxConcurrentJobs int
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	URL     string
	Timeout time.Duration
}

// ReportsConfig holds reports configuration
type ReportsConfig struct {
	Directory string
}

// Load loads configuration from environment variables
// It first tries to load from .env file, then from system environment
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{}

	// Load database configuration
	config.Database = DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", "my_aibo_app"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Load server configuration
	config.Server = ServerConfig{
		Host: getEnv("SERVER_HOST", "0.0.0.0"),
		Port: getEnvAsInt("SERVER_PORT", 8080),
	}

	// Load application configuration
	config.App = AppConfig{
		Environment: getEnv("APP_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	// Load scheduler configuration
	config.Scheduler = SchedulerConfig{
		Enabled:           getEnvAsBool("SCHEDULER_ENABLED", true),
		MaxConcurrentJobs: getEnvAsInt("MAX_CONCURRENT_JOBS", 10),
	}

	// Load health check configuration
	healthCheckTimeout, err := time.ParseDuration(getEnv("HEALTH_CHECK_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid HEALTH_CHECK_TIMEOUT: %w", err)
	}

	config.HealthCheck = HealthCheckConfig{
		URL:     getEnv("HEALTH_CHECK_URL", "https://httpbin.org/status/200"),
		Timeout: healthCheckTimeout,
	}

	// Load reports configuration
	config.Reports = ReportsConfig{
		Directory: getEnv("REPORTS_DIR", "./reports"),
	}

	return config, nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// SetupLogger configures the logger based on configuration
func (c *Config) SetupLogger() {
	// Set log level
	level, err := logrus.ParseLevel(c.App.LogLevel)
	if err != nil {
		logrus.Warnf("Invalid log level '%s', using 'info'", c.App.LogLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set log format based on environment
	if c.App.Environment == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
