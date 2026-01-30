package config

import (
	"fmt"
	"os"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	DatabasePath       string
	CookieDomain       string
	CookieSecure       bool
	ServerPort         string
	FrontendURL        string
}

func Load() (*Config, error) {
	cfg := &Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		DatabasePath:       getEnvOrDefault("DATABASE_PATH", "bankodad.db"),
		CookieDomain:       os.Getenv("COOKIE_DOMAIN"),
		CookieSecure:       os.Getenv("COOKIE_SECURE") != "false",
		ServerPort:         getEnvOrDefault("SERVER_PORT", "8001"),
		FrontendURL:        getEnvOrDefault("FRONTEND_URL", "http://localhost:8000"),
	}

	if cfg.GoogleClientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable is required")
	}
	if cfg.GoogleClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET environment variable is required")
	}
	if cfg.GoogleRedirectURL == "" {
		cfg.GoogleRedirectURL = fmt.Sprintf("http://localhost:%s/api/auth/google/callback", cfg.ServerPort)
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
