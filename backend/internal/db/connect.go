package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"

	"github.com/Dysar/url-crawler/backend/internal/config"
)

func NewMySQLConnection(cfg config.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	
	// Optimize connection pool settings
	db.SetMaxOpenConns(25)                 // Max open connections
	db.SetMaxIdleConns(5)                  // Max idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection max lifetime
	db.SetConnMaxIdleTime(10 * time.Minute) // Connection max idle time
	
	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	return db, nil
}


