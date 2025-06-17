# Job Scheduler Microservice

A production-ready job scheduler microservice built with Go and PostgreSQL, featuring cron-based scheduling, RESTful API, and clean architecture.

## 🚀 Features

- **RESTful API**: Complete CRUD operations for job management
- **Cron-based Scheduling**: Flexible job scheduling with cron expressions
- **Multiple Job Types**: Email notifications, data processing, reports, health checks
- **PostgreSQL Integration**: Robust data persistence with JSONB configuration
- **Clean Architecture**: Separation of concerns with dependency injection
- **Docker Support**: Containerized deployment ready
- **Health Monitoring**: Built-in health checks and structured logging

## 🛠️ Technology Stack

- **Backend**: Go 1.18+
- **Database**: PostgreSQL 13+
- **Web Framework**: Gin
- **ORM**: GORM
- **Scheduling**: robfig/cron
- **Containerization**: Docker & Docker Compose

## 📋 Prerequisites

- Go 1.18 or later
- PostgreSQL 13 or later
- Docker and Docker Compose (optional)

## ⚡ Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd job-scheduler

# Start with Docker Compose
docker-compose up --build

# Test the API
curl http://localhost:8080/api/v1/health
```

### Option 2: Manual Setup

1. **Install Dependencies**
   ```bash
   go mod download
   ```

2. **Setup Database**
   ```bash
   # Create PostgreSQL database
   createdb jobscheduler_db

   # Create user
   psql -c "CREATE USER jobscheduler WITH PASSWORD 'password123';"
   psql -c "GRANT ALL PRIVILEGES ON DATABASE jobscheduler_db TO jobscheduler;"
   ```

3. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your database settings
   ```

4. **Run the Application**
   ```bash
   go run cmd/server/main.go
   ```

## 🔧 Configuration

Create a `.env` file:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=jobscheduler
DB_PASSWORD=password123
DB_NAME=jobscheduler_db

# Server Configuration
SERVER_PORT=8080
APP_ENV=development
LOG_LEVEL=info

# Scheduler Configuration
SCHEDULER_ENABLED=true
MAX_CONCURRENT_JOBS=10
```

## 📡 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/jobs` | List all jobs |
| GET | `/api/v1/jobs/{id}` | Get job by ID |
| POST | `/api/v1/jobs` | Create new job |
| PUT | `/api/v1/jobs/{id}` | Update job |
| DELETE | `/api/v1/jobs/{id}` | Delete job |

### Example: Create a Job

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Email Report",
    "description": "Send daily summary",
    "schedule": "0 9 * * *",
    "job_type": "email_notification",
    "config": {
      "recipient": "admin@example.com",
      "subject": "Daily Summary"
    }
  }'
```

## 🏗️ Architecture

Clean Architecture with separation of concerns:

```
cmd/server/          # Application entry point
├── internal/
│   ├── handlers/    # HTTP handlers
│   ├── services/    # Business logic
│   ├── repositories/# Data access
│   ├── models/      # Domain models
│   └── scheduler/   # Job scheduling
├── pkg/database/    # Database utilities
└── migrations/      # SQL migrations
```

## 📊 Job Types

1. **Email Notification**: Send emails with configurable content
2. **Data Processing**: Execute data transformation tasks
3. **Report Generation**: Generate reports in various formats
4. **Health Check**: Monitor external services

## 🔄 Cron Schedule Examples

- `0 9 * * *` - Daily at 9:00 AM
- `*/5 * * * *` - Every 5 minutes
- `0 0 * * 0` - Weekly on Sunday at midnight
- `0 9 1 * *` - Monthly on the 1st at 9:00 AM

## 🧪 Testing

```bash
# Run unit tests
go test ./tests/...

# Run API integration tests
./test_api.sh
```

## 🚀 Deployment

```bash
# Docker deployment
docker-compose up --build -d

# View logs
docker-compose logs -f
```

## 🔍 Monitoring

- **Health Check**: `/api/v1/health`
- **Structured Logging**: JSON formatted logs
- **Database Health**: Connection monitoring
- **Scheduler Status**: Job execution tracking

## 🆘 Troubleshooting

**Database Connection Issues:**
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Test connection
psql -h localhost -U jobscheduler -d jobscheduler_db
```

**Port Conflicts:**
```bash
# Find process using port
lsof -i :8080

# Change port in .env
SERVER_PORT=8081
```

For detailed setup and troubleshooting, see `BEGINNER_GUIDE.md`.

---

**Built with Go and PostgreSQL**
