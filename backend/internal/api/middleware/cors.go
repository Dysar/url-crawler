package middleware

import (
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
)

// CORS enables Cross-Origin Resource Sharing for the API.
// For the test task we allow configurable origin via FRONTEND_ORIGIN; defaults to *.
func CORS() gin.HandlerFunc {
    allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
    if allowedOrigin == "" {
        allowedOrigin = "*"
    }
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        c.Next()
    }
}


