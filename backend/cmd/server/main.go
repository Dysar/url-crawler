package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Dysar/url-crawler/backend/internal/api"
	"github.com/Dysar/url-crawler/backend/internal/config"
	"github.com/Dysar/url-crawler/backend/internal/crawler"
	"github.com/Dysar/url-crawler/backend/internal/db"
	"github.com/Dysar/url-crawler/backend/internal/repository"
	"github.com/Dysar/url-crawler/backend/internal/service"
)

func main() {
	cfg := config.Load()

	conn, err := db.NewMySQLConnection(cfg)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
	}
	if conn != nil {
		if sqlDB, err2 := conn.DB(); err2 == nil {
			defer sqlDB.Close()
		}
		log.Printf("database connection established")
	}

	r := gin.New()
	r.Use(gin.Recovery())

	deps := api.Deps{}
	if conn != nil {
		deps.URLRepo = repository.NewURLRepository(conn)
		deps.JobRepo = repository.NewJobRepository(conn)
		resRepo := repository.NewResultRepository(conn)
		deps.ResultRepo = resRepo
		cr := crawler.New(crawler.HTTPClient(30 * time.Second))
		deps.JobService = service.NewJobService(deps.JobRepo, resRepo, deps.URLRepo, cr)
	}
	api.RegisterRoutes(r, cfg, deps)

	addr := ":" + cfg.APIPort
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}
