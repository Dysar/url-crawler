package middleware

import (
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v5"

    "github.com/Dysar/url-crawler/backend/internal/config"
)

type loginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func AuthLoginHandler(cfg config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req loginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
            return
        }

        // Simple single-user auth for test task, credentials via env or defaults
        adminUser := os.Getenv("ADMIN_USERNAME")
        if adminUser == "" {
            adminUser = "admin"
        }
        adminPass := os.Getenv("ADMIN_PASSWORD")
        if adminPass == "" {
            adminPass = "password"
        }

        if req.Username != adminUser || req.Password != adminPass {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }

        claims := jwt.MapClaims{
            "sub": req.Username,
            "role": "admin",
            "exp": time.Now().Add(24 * time.Hour).Unix(),
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        signed, err := token.SignedString([]byte(cfg.JWTSecret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"token": signed})
    }
}

func JWTAuth(cfg config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
            return
        }
        tokenStr := auth[7:]
        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrTokenUnverifiable
            }
            return []byte(cfg.JWTSecret), nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        c.Next()
    }
}



