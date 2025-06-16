# Job Scheduler Microservice

A comprehensive job scheduling microservice built with Go and PostgreSQL for interview assessment. This service provides a REST API for managing scheduled jobs and includes a background scheduler that executes jobs based on cron expressions.

## 🚀 Features

- **REST API**: Complete CRUD operations for job management
- **Background Scheduler**: Cron-based job execution with concurrency control
- **Multiple Job Types**: Email notifications, data processing, report generation, and health checks
- **PostgreSQL Integration**: Robust data persistence with GORM
- **Docker Support**: Easy local development and deployment
- **Comprehensive Logging**: Structured logging with logrus
- **Graceful Shutdown**: Proper cleanup of resources
- **Health Checks**: Built-in health monitoring endpoints

## 📋 Prerequisites

Before running this application, ensure you have the following installed:

- **Go 1.21 or later**: [Download Go](https://golang.org/dl/)
- **PostgreSQL 12 or later**: [Download PostgreSQL](https://www.postgresql.org/download/)
- **Docker & Docker Compose** (optional): [Download Docker](https://www.docker.com/get-started)
- **Git**: [Download Git](https://git-scm.com/downloads)

## 🛠️ Quick Start

### Option 1: Using Docker (Recommended)

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd job-scheduler
   ```

2. **Start the application**:
   ```bash
   docker-compose up --build
   ```

3. **Verify the application is running**:
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

### Option 2: Manual Setup

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd job-scheduler
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Setup PostgreSQL database**:
   ```sql
   CREATE DATABASE jobscheduler_db;
   CREATE USER jobscheduler WITH PASSWORD 'password123';
   GRANT ALL PRIVILEGES ON DATABASE jobscheduler_db TO jobscheduler;
   ```

4. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

5. **Run the application**:
   ```bash
   go run cmd/server/main.go
   ```

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### Health Check
```http
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "services": {
    "database": {"status": "healthy", "response_time_ms": 5},
    "scheduler": {"status": "healthy", "is_running": true, "scheduled_jobs": 4}
  }
}
```

#### List Jobs
```http
GET /jobs?page=1&limit=10
```

**Response**:
```json
{
  "jobs": [...],
  "total_count": 25,
  "page": 1,
  "limit": 10,
  "total_pages": 3
}
```

#### Get Job by ID
```http
GET /jobs/{id}
```

#### Create Job
```http
POST /jobs
Content-Type: application/json

{
  "name": "Daily Backup",
  "description": "Backup database daily at midnight",
  "schedule": "0 0 * * *",
  "job_type": "data_processing",
  "config": {
    "processing_time_seconds": 10,
    "operation": "backup"
  },
  "is_active": true
}
```

#### Update Job
```http
PUT /jobs/{id}
Content-Type: application/json

{
  "name": "Updated Job Name",
  "is_active": false
}
```

#### Delete Job
```http
DELETE /jobs/{id}
```

### Job Types

1. **email_notification**: Simulates sending emails
2. **data_processing**: Simulates data processing tasks
3. **report_generation**: Creates text reports in the reports directory
4. **health_check**: Performs HTTP health checks

### Cron Schedule Format

The service uses standard cron expressions:
```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, Sunday = 0 or 7)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

**Examples**:
- `0 9 * * *` - Every day at 9:00 AM
- `*/5 * * * *` - Every 5 minutes
- `0 0 * * 1` - Every Monday at midnight
- `30 14 * * 1-5` - Every weekday at 2:30 PM

## 🧪 Testing

### Manual Testing with curl

1. **Create a test job**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/jobs \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Test Job",
       "description": "A test job for demonstration",
       "schedule": "*/1 * * * *",
       "job_type": "email_notification",
       "config": {
         "recipient": "test@example.com",
         "subject": "Test Email"
       }
     }'
   ```

2. **List all jobs**:
   ```bash
   curl http://localhost:8080/api/v1/jobs
   ```

3. **Get specific job**:
   ```bash
   curl http://localhost:8080/api/v1/jobs/{job-id}
   ```

4. **Check health**:
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

### Running Unit Tests

```bash
go test ./...
```

## 🏗️ Architecture

The application follows clean architecture principles:

```
cmd/server/          # Application entry point
internal/
├── config/          # Configuration management
├── models/          # Data models and DTOs
├── repositories/    # Data access layer
├── services/        # Business logic layer
├── handlers/        # HTTP handlers
└── scheduler/       # Background job scheduler
pkg/database/        # Database utilities
migrations/          # SQL migration files
```

### Key Components

- **Models**: Define data structures for jobs and executions
- **Repositories**: Handle database operations using GORM
- **Services**: Implement business logic and validation
- **Handlers**: Process HTTP requests and responses
- **Scheduler**: Manage cron-based job execution
- **Executor**: Handle individual job execution with concurrency control

## 🔧 Configuration

Environment variables (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | Database user | jobscheduler |
| `DB_PASSWORD` | Database password | password123 |
| `DB_NAME` | Database name | jobscheduler_db |
| `SERVER_PORT` | HTTP server port | 8080 |
| `LOG_LEVEL` | Logging level | info |
| `SCHEDULER_ENABLED` | Enable job scheduler | true |
| `MAX_CONCURRENT_JOBS` | Max concurrent job executions | 10 |

## 📊 Monitoring

### Logs

The application provides structured logging with the following levels:
- `error`: Critical errors
- `warn`: Warning messages
- `info`: General information
- `debug`: Detailed debugging information

### Health Checks

Monitor application health via `/api/v1/health` endpoint, which checks:
- Database connectivity
- Scheduler status
- Overall system health

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Failed**:
   - Verify PostgreSQL is running
   - Check database credentials in `.env`
   - Ensure database exists

2. **Port Already in Use**:
   - Change `SERVER_PORT` in `.env`
   - Kill process using the port: `lsof -ti:8080 | xargs kill`

3. **Jobs Not Executing**:
   - Check if scheduler is enabled (`SCHEDULER_ENABLED=true`)
   - Verify cron expressions are valid
   - Check job `is_active` status

4. **Docker Issues**:
   - Ensure Docker daemon is running
   - Try `docker-compose down && docker-compose up --build`

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
```

## 🔄 Development

### Adding New Job Types

1. Create executor in `internal/services/job_types.go`
2. Implement `JobExecutor` interface
3. Register in `NewJobExecutor` function
4. Add to `JobType` enum in models

### Database Migrations

Migrations are automatically applied on startup. For manual migration:
```bash
psql -h localhost -U jobscheduler -d jobscheduler_db -f migrations/001_create_jobs_table.sql
```

## 📝 License

This project is created for interview assessment purposes.

## 🤝 Contributing

This is an assessment project. For questions or clarifications, please contact the development team.
