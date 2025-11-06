package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Dysar/url-crawler/backend/internal/api/handlers"
	"github.com/Dysar/url-crawler/backend/internal/api/middleware"
	"github.com/Dysar/url-crawler/backend/internal/config"
	"github.com/Dysar/url-crawler/backend/internal/repository"
	"github.com/Dysar/url-crawler/backend/internal/service"
)

type responseEnvelope struct {
	Data    any     `json:"data"`
	Error   *string `json:"error"`
	Message string  `json:"message"`
}

// RegisterRoutes wires routes with dependencies.
func RegisterRoutes(r *gin.Engine, cfg config.Config, deps Deps) {
	// Global middlewares
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, responseEnvelope{Data: gin.H{"ok": true}, Error: nil, Message: "OK"})
	})

	api := r.Group("/api/v1")

	api.POST("/auth/login", middleware.AuthLoginHandler(cfg))

	secured := api.Group("")
	secured.Use(middleware.JWTAuth(cfg))
	{
		// Placeholders for future endpoints
		secured.GET("/me", func(c *gin.Context) {
			c.JSON(http.StatusOK, responseEnvelope{Data: gin.H{"role": "admin"}, Error: nil, Message: "Success"})
		})

		// URL management
		urlHandlers := handlers.NewURLHandlers(deps.URLRepo)
		secured.POST("/urls", urlHandlers.CreateURL)
		secured.GET("/urls", urlHandlers.ListURLs)

		// jobs
		jobHandlers := handlers.NewJobHandlers(deps.URLRepo, deps.JobRepo, deps.JobService)
		secured.POST("/jobs/start", jobHandlers.Start)
		secured.POST("/jobs/stop", jobHandlers.Stop)
		secured.GET("/jobs/:id/status", jobHandlers.Status)

		// results
		resultHandlers := handlers.NewResultHandlers(deps.ResultRepo)
		secured.GET("/results/:id", resultHandlers.GetByURLID)
	}
}

// Deps contains runtime dependencies for handlers.
type Deps struct {
	URLRepo    repository.URLRepository
	JobRepo    repository.JobRepository
	ResultRepo repository.ResultRepository
	JobService *service.JobService
}
