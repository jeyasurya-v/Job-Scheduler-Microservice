-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schedule VARCHAR(100) NOT NULL,
    job_type VARCHAR(50) NOT NULL,
    config JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_jobs_is_active ON jobs(is_active);
CREATE INDEX IF NOT EXISTS idx_jobs_job_type ON jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_jobs_updated_at 
    BEFORE UPDATE ON jobs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample jobs for testing (using ON CONFLICT to avoid duplicate key errors)
INSERT INTO jobs (name, description, schedule, job_type, config, is_active) VALUES
(
    'Daily Email Notification',
    'Send daily summary email to administrators',
    '0 9 * * *',
    'email_notification',
    '{"recipient": "admin@example.com", "subject": "Daily Summary", "body": "Your daily summary is ready."}',
    true
),
(
    'Hourly Data Processing',
    'Process incoming data every hour',
    '0 * * * *',
    'data_processing',
    '{"processing_time_seconds": 3, "data_size": "500KB", "operation": "transform"}',
    true
),
(
    'Weekly Report Generation',
    'Generate weekly reports every Monday at 8 AM',
    '0 8 * * 1',
    'report_generation',
    '{"report_type": "weekly_summary", "format": "txt", "include_charts": false}',
    true
),
(
    'Health Check Every 10 Minutes',
    'Check system health every 10 minutes',
    '0,10,20,30,40,50 * * * *',
    'health_check',
    '{"url": "https://httpbin.org/status/200", "timeout_seconds": 30, "expected_status": 200}',
    true
)
ON CONFLICT (name) DO NOTHING;
