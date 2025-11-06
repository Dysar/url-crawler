package config

import (
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	APIPort    string
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func Load() Config {
	return Config{
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "3306"),
		DBUser:     getenv("DB_USER", "root"),
		DBPassword: getenv("DB_PASSWORD", "root"),
		DBName:     getenv("DB_NAME", "url_crawler"),
		JWTSecret:  getenv("JWT_SECRET", "dev-secret-change"),
		APIPort:    getenv("API_PORT", "8080"),
	}
}
