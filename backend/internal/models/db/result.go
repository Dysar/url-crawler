package db

import "time"

type CrawlResult struct {
    ID                       int64     `gorm:"column:id;primaryKey;autoIncrement"`
    JobID                    int64     `gorm:"column:job_id;not null"`
    URLID                    int64     `gorm:"column:url_id;not null"`
    HTMLVersion              *string   `gorm:"column:html_version;size:50"`
    Title                    *string   `gorm:"column:title;size:500"`
    HeadingsH1               int       `gorm:"column:headings_h1"`
    HeadingsH2               int       `gorm:"column:headings_h2"`
    HeadingsH3               int       `gorm:"column:headings_h3"`
    HeadingsH4               int       `gorm:"column:headings_h4"`
    HeadingsH5               int       `gorm:"column:headings_h5"`
    HeadingsH6               int       `gorm:"column:headings_h6"`
    InternalLinksCount       int       `gorm:"column:internal_links_count"`
    ExternalLinksCount       int       `gorm:"column:external_links_count"`
    InaccessibleLinksCount   int       `gorm:"column:inaccessible_links_count"`
    HasLoginForm             bool      `gorm:"column:has_login_form"`
    CreatedAt                time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (CrawlResult) TableName() string { return "crawl_results" }


