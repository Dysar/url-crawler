-- Initial schema for URL Crawler

CREATE TABLE IF NOT EXISTS urls (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    url VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_url (url(255))
);

CREATE TABLE IF NOT EXISTS crawl_jobs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    url_id BIGINT NOT NULL,
    status ENUM('queued', 'running', 'completed', 'failed') DEFAULT 'queued',
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_url_id (url_id)
);

CREATE TABLE IF NOT EXISTS crawl_results (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    job_id BIGINT NOT NULL,
    url_id BIGINT NOT NULL,
    html_version VARCHAR(50) NULL,
    title VARCHAR(500) NULL,
    headings_h1 INT DEFAULT 0,
    headings_h2 INT DEFAULT 0,
    headings_h3 INT DEFAULT 0,
    headings_h4 INT DEFAULT 0,
    headings_h5 INT DEFAULT 0,
    headings_h6 INT DEFAULT 0,
    internal_links_count INT DEFAULT 0,
    external_links_count INT DEFAULT 0,
    inaccessible_links_count INT DEFAULT 0,
    has_login_form BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES crawl_jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
    UNIQUE KEY unique_job_result (job_id)
);


