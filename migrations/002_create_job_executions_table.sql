-- Create job_executions table
CREATE TABLE IF NOT EXISTS job_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    execution_duration BIGINT, -- Duration in milliseconds
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_job_executions_job_id ON job_executions(job_id);
CREATE INDEX IF NOT EXISTS idx_job_executions_status ON job_executions(status);
CREATE INDEX IF NOT EXISTS idx_job_executions_started_at ON job_executions(started_at);
CREATE INDEX IF NOT EXISTS idx_job_executions_created_at ON job_executions(created_at);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_job_executions_job_id_started_at ON job_executions(job_id, started_at DESC);

-- Add check constraint for status values
ALTER TABLE job_executions 
ADD CONSTRAINT chk_job_executions_status 
CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'));

-- Add check constraint for execution_duration (must be positive if not null)
ALTER TABLE job_executions 
ADD CONSTRAINT chk_job_executions_duration 
CHECK (execution_duration IS NULL OR execution_duration >= 0);
