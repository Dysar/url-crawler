package db

import "time"

type URL struct {
    ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
    URL       string    `gorm:"column:url;size:2048;not null"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (URL) TableName() string { return "urls" }


