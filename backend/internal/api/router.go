package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Dysar/url-crawler/backend/internal/api/handlers"
	"github.com/Dysar/url-crawler/backend/internal/api/middleware"
	"github.com/Dysar/url-crawler/backend/internal/config"
	"github.com/Dysar/url-crawler/backend/internal/service"
	"github.com/gin-contrib/cors"
)

type responseEnvelope struct {
	Data    any     `json:"data"`
	Error   *string `json:"error"`
	Message string  `json:"message"`
}

// RegisterRoutes wires routes with dependencies.
func RegisterRoutes(r *gin.Engine, cfg config.Config, deps Deps) {
	// Global middlewares - configure CORS to allow Authorization header
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Referer", "User-Agent"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, responseEnvelope{Data: gin.H{"ok": true}, Error: nil, Message: "OK"})
	})

	api := r.Group("/api/v1")

	api.POST("/auth/login", middleware.AuthLoginHandler(cfg))

	secured := api.Group("")
	secured.Use(middleware.JWTAuth(cfg))
	{
		// URL management
		urlHandlers := handlers.NewURLHandlers(deps.URLService)
		secured.POST("/urls", urlHandlers.CreateURL)
		secured.GET("/urls", urlHandlers.ListURLs)

		// jobs
		jobHandlers := handlers.NewJobHandlers(deps.JobService)
		secured.POST("/jobs/start", jobHandlers.Start)
		secured.POST("/jobs/stop", jobHandlers.Stop)
		secured.GET("/jobs/:id/status", jobHandlers.Status)

		// results
		resultHandlers := handlers.NewResultHandlers(deps.ResultService)
		secured.GET("/results/:id", resultHandlers.GetByURLID)
	}
}

// Deps contains runtime dependencies for handlers.
type Deps struct {
	URLService    *service.URLService
	JobService    *service.JobService
	ResultService *service.ResultService
}
