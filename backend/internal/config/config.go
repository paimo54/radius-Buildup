package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret      string
	AgentJWTSecret string

	// App
	AppName string
	AppURL  string
	TZ      string
}

// Load reads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Try loading .env from parent directory (shared with Next.js)
	_ = godotenv.Load("../.env")
	// Also try local .env
	_ = godotenv.Load(".env")

	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "3306"))

	cfg := &Config{
		Port: getEnv("GO_PORT", "8080"),
		Env:  getEnv("GO_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "radius_buildup"),

		JWTSecret:      getEnv("NEXTAUTH_SECRET", ""),
		AgentJWTSecret: getEnv("AGENT_JWT_SECRET", ""),

		AppName: getEnv("NEXT_PUBLIC_APP_NAME", "Radius Buildup"),
		AppURL:  getEnv("NEXT_PUBLIC_APP_URL", "http://localhost:3000"),
		TZ:      getEnv("TZ", "Asia/Jakarta"),
	}

	return cfg, nil
}

// DSN returns the MySQL connection string for GORM
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
