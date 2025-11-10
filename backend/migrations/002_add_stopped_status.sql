-- Add 'stopped' status to crawl_jobs ENUM
ALTER TABLE crawl_jobs MODIFY COLUMN status ENUM('queued', 'running', 'done', 'error', 'stopped') DEFAULT 'queued';

