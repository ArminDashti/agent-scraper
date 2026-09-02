package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	JWTSecret       string
	DefaultUsername string
	DefaultPassword string
	CORSOrigins     []string
	CronSchedule    string
	TargetAPIURL    string
}

func Load() Config {
	return Config{
		HTTPAddr:        getenv("HTTP_ADDR", "127.0.0.1:8196"),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://agentscraper:agentscraper@127.0.0.1:5456/agentscraper?sslmode=disable"),
		JWTSecret:       getenv("JWT_SECRET", "change-me-in-production"),
		DefaultUsername: getenv("DEFAULT_USERNAME", "armin"),
		DefaultPassword: getenv("DEFAULT_PASSWORD", "dopadopa123"),
		CORSOrigins:     splitCSV(getenv("CORS_ORIGINS", "http://127.0.0.1:5196,http://localhost:5196")),
		CronSchedule:    getenv("CRON_SCHEDULE", "0 2 * * *"),
		TargetAPIURL:    getenv("TARGET_API_URL", ""),
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
