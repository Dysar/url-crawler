package models

type CreateURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type URLResponse struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type URLListResponse struct {
	Data  []URLResponse `json:"data"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

// ResultResponse represents the API response model for a crawl result
type ResultResponse struct {
	ID                     int64   `json:"id"`
	URLID                  int64   `json:"url_id"`
	HTMLVersion            *string `json:"html_version"`
	Title                  *string `json:"title"`
	HeadingsH1             int     `json:"headings_h1"`
	HeadingsH2             int     `json:"headings_h2"`
	HeadingsH3             int     `json:"headings_h3"`
	HeadingsH4             int     `json:"headings_h4"`
	HeadingsH5             int     `json:"headings_h5"`
	HeadingsH6             int     `json:"headings_h6"`
	InternalLinksCount     int     `json:"internal_links_count"`
	ExternalLinksCount     int     `json:"external_links_count"`
	InaccessibleLinksCount int     `json:"inaccessible_links_count"`
	HasLoginForm           bool    `json:"has_login_form"`
}

// Jobs API response types

type JobStartResponse struct {
	URLID int64 `json:"url_id"`
	JobID int64 `json:"job_id"`
}

type JobStatusResponse struct {
	ID          int64          `json:"id"`
	Status      CrawlJobStatus `json:"status"`
	Error       *string        `json:"error"`
	StartedAt   *string        `json:"started_at,omitempty"`
	CompletedAt *string        `json:"completed_at,omitempty"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type JobsStoppedItem struct {
	URLID int64 `json:"url_id"`
	JobID int64 `json:"job_id"`
}
type JobsStoppedResponse []JobsStoppedItem
