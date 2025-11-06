package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/yourname/url-crawler/backend/internal/api"
	"github.com/yourname/url-crawler/backend/internal/config"
	"github.com/yourname/url-crawler/backend/internal/db"
)

func main() {
	cfg := config.Load()

	if conn, err := db.NewMySQLConnection(cfg); err != nil {
		log.Printf("failed to connect to database: %v", err)
	} else {
		sqlDB, _ := conn.DB()
		defer sqlDB.Close()
		log.Printf("database connection established")
	}

	r := gin.New()
	r.Use(gin.Recovery())

	api.RegisterRoutes(r, cfg)

	addr := ":" + cfg.APIPort
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}
