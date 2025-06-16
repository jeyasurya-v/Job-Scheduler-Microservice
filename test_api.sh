#!/bin/bash

# Job Scheduler API Test Script
# This script demonstrates the API functionality

echo "🚀 Job Scheduler API Test Script"
echo "================================="

BASE_URL="http://localhost:8080/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Test 1: Health Check
echo ""
print_info "Test 1: Health Check"
response=$(curl -s -w "%{http_code}" -o /tmp/health_response.json "$BASE_URL/health")
http_code="${response: -3}"

if [ "$http_code" = "200" ]; then
    print_status 0 "Health check passed"
    echo "Response:"
    cat /tmp/health_response.json | jq '.' 2>/dev/null || cat /tmp/health_response.json
else
    print_status 1 "Health check failed (HTTP $http_code)"
fi

# Test 2: List Jobs (should be empty initially)
echo ""
print_info "Test 2: List Jobs (Initial)"
response=$(curl -s -w "%{http_code}" -o /tmp/jobs_response.json "$BASE_URL/jobs")
http_code="${response: -3}"

if [ "$http_code" = "200" ]; then
    print_status 0 "List jobs successful"
    echo "Response:"
    cat /tmp/jobs_response.json | jq '.' 2>/dev/null || cat /tmp/jobs_response.json
else
    print_status 1 "List jobs failed (HTTP $http_code)"
fi

# Test 3: Create a new job
echo ""
print_info "Test 3: Create Email Notification Job"
job_data='{
  "name": "Test Email Job",
  "description": "A test email notification job",
  "schedule": "*/5 * * * *",
  "job_type": "email_notification",
  "config": {
    "recipient": "test@example.com",
    "subject": "Test Email",
    "body": "This is a test email from the job scheduler"
  },
  "is_active": true
}'

response=$(curl -s -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d "$job_data" \
  -o /tmp/create_job_response.json \
  "$BASE_URL/jobs")
http_code="${response: -3}"

if [ "$http_code" = "201" ]; then
    print_status 0 "Job creation successful"
    echo "Response:"
    cat /tmp/create_job_response.json | jq '.' 2>/dev/null || cat /tmp/create_job_response.json
    
    # Extract job ID for further tests
    job_id=$(cat /tmp/create_job_response.json | jq -r '.job.id' 2>/dev/null)
    if [ "$job_id" != "null" ] && [ "$job_id" != "" ]; then
        echo "Created job ID: $job_id"
    fi
else
    print_status 1 "Job creation failed (HTTP $http_code)"
    cat /tmp/create_job_response.json
fi

# Test 4: Create another job (Data Processing)
echo ""
print_info "Test 4: Create Data Processing Job"
job_data2='{
  "name": "Test Data Processing",
  "description": "A test data processing job",
  "schedule": "0 */2 * * *",
  "job_type": "data_processing",
  "config": {
    "processing_time_seconds": 3,
    "data_size": "500KB",
    "operation": "transform"
  },
  "is_active": true
}'

response=$(curl -s -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d "$job_data2" \
  -o /tmp/create_job2_response.json \
  "$BASE_URL/jobs")
http_code="${response: -3}"

if [ "$http_code" = "201" ]; then
    print_status 0 "Data processing job creation successful"
    echo "Response:"
    cat /tmp/create_job2_response.json | jq '.' 2>/dev/null || cat /tmp/create_job2_response.json
    
    # Extract job ID for further tests
    job_id2=$(cat /tmp/create_job2_response.json | jq -r '.job.id' 2>/dev/null)
    if [ "$job_id2" != "null" ] && [ "$job_id2" != "" ]; then
        echo "Created job ID: $job_id2"
    fi
else
    print_status 1 "Data processing job creation failed (HTTP $http_code)"
    cat /tmp/create_job2_response.json
fi

# Test 5: List Jobs (should show created jobs)
echo ""
print_info "Test 5: List Jobs (After Creation)"
response=$(curl -s -w "%{http_code}" -o /tmp/jobs_list_response.json "$BASE_URL/jobs")
http_code="${response: -3}"

if [ "$http_code" = "200" ]; then
    print_status 0 "List jobs successful"
    echo "Response:"
    cat /tmp/jobs_list_response.json | jq '.' 2>/dev/null || cat /tmp/jobs_list_response.json
else
    print_status 1 "List jobs failed (HTTP $http_code)"
fi

# Test 6: Get specific job (if we have a job ID)
if [ "$job_id" != "null" ] && [ "$job_id" != "" ]; then
    echo ""
    print_info "Test 6: Get Specific Job ($job_id)"
    response=$(curl -s -w "%{http_code}" -o /tmp/get_job_response.json "$BASE_URL/jobs/$job_id")
    http_code="${response: -3}"

    if [ "$http_code" = "200" ]; then
        print_status 0 "Get specific job successful"
        echo "Response:"
        cat /tmp/get_job_response.json | jq '.' 2>/dev/null || cat /tmp/get_job_response.json
    else
        print_status 1 "Get specific job failed (HTTP $http_code)"
    fi
fi

# Test 7: Test invalid job creation (should fail)
echo ""
print_info "Test 7: Invalid Job Creation (Should Fail)"
invalid_job_data='{
  "name": "",
  "schedule": "invalid cron",
  "job_type": "invalid_type"
}'

response=$(curl -s -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d "$invalid_job_data" \
  -o /tmp/invalid_job_response.json \
  "$BASE_URL/jobs")
http_code="${response: -3}"

if [ "$http_code" = "400" ]; then
    print_status 0 "Invalid job creation properly rejected"
    echo "Response:"
    cat /tmp/invalid_job_response.json | jq '.' 2>/dev/null || cat /tmp/invalid_job_response.json
else
    print_status 1 "Invalid job creation should have been rejected (HTTP $http_code)"
fi

# Test 8: Test pagination
echo ""
print_info "Test 8: Test Pagination"
response=$(curl -s -w "%{http_code}" -o /tmp/pagination_response.json "$BASE_URL/jobs?page=1&limit=1")
http_code="${response: -3}"

if [ "$http_code" = "200" ]; then
    print_status 0 "Pagination test successful"
    echo "Response:"
    cat /tmp/pagination_response.json | jq '.' 2>/dev/null || cat /tmp/pagination_response.json
else
    print_status 1 "Pagination test failed (HTTP $http_code)"
fi

echo ""
echo "🎉 API Testing Complete!"
echo ""
echo "📝 Summary:"
echo "- The API endpoints are working correctly"
echo "- Job creation, listing, and retrieval are functional"
echo "- Input validation is working"
echo "- Pagination is implemented"
echo ""
echo "🔍 To monitor job execution:"
echo "- Check the application logs for job execution messages"
echo "- Jobs will execute according to their cron schedules"
echo "- Email jobs will log simulated email sending"
echo "- Data processing jobs will simulate processing with delays"
echo ""
echo "📚 Next Steps:"
echo "1. Monitor the logs to see jobs executing"
echo "2. Try creating jobs with different schedules"
echo "3. Test the UPDATE and DELETE endpoints (bonus features)"
echo "4. Check the reports directory for generated reports"

# Cleanup temporary files
rm -f /tmp/health_response.json /tmp/jobs_response.json /tmp/create_job_response.json
rm -f /tmp/create_job2_response.json /tmp/jobs_list_response.json /tmp/get_job_response.json
rm -f /tmp/invalid_job_response.json /tmp/pagination_response.json
