package api

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/yourname/url-crawler/backend/internal/api/middleware"
    "github.com/yourname/url-crawler/backend/internal/config"
)

type responseEnvelope struct {
    Data    any    `json:"data"`
    Error   *string `json:"error"`
    Message string `json:"message"`
}

func RegisterRoutes(r *gin.Engine, cfg config.Config) {
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
    }
}


