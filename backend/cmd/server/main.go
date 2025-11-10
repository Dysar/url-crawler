package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()
	log.Printf("database connection established")

	r := gin.New()
	r.Use(gin.Recovery())

	// Create repositories
	urlRepo := repository.NewURLRepository(conn)
	jobRepo := repository.NewJobRepository(conn)
	resultRepo := repository.NewResultRepository(conn)

	// Create services
	urlService, err := service.NewURLService(urlRepo)
	if err != nil {
		log.Fatalf("failed to create URL service: %v", err)
	}

	resultService, err := service.NewResultService(resultRepo)
	if err != nil {
		log.Fatalf("failed to create result service: %v", err)
	}

	cr := crawler.New(crawler.HTTPClient(30 * time.Second))
	jobService, err := service.NewJobService(jobRepo, resultRepo, urlRepo, cr)
	if err != nil {
		log.Fatalf("failed to create job service: %v", err)
	}

	deps := api.Deps{
		URLService:    urlService,
		JobService:    jobService,
		ResultService: resultService,
	}
	api.RegisterRoutes(r, cfg, deps)

	// Create HTTP server
	addr := ":" + cfg.APIPort
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// Gracefully shutdown job service workers
	log.Println("shutting down job service workers...")
	jobService.Shutdown()

	// Gracefully shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
