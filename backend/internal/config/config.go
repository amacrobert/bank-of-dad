package config

import (
	"encoding/base64"
	"fmt"
	"os"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	DatabaseURL        string
	JWTSecret          []byte
	ServerPort         string
	FrontendURL        string
}

func Load() (*Config, error) {
	cfg := &Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		DatabaseURL:        getEnvOrDefault("DATABASE_URL", "postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable"),
		ServerPort:         getEnvOrDefault("SERVER_PORT", "8001"),
		FrontendURL:        getEnvOrDefault("FRONTEND_URL", "http://localhost:8000"),
	}

	// Decode JWT secret
	jwtSecretB64 := os.Getenv("JWT_SECRET")
	if jwtSecretB64 == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	jwtSecret, err := base64.StdEncoding.DecodeString(jwtSecretB64)
	if err != nil {
		return nil, fmt.Errorf("JWT_SECRET must be valid base64: %w", err)
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes (got %d)", len(jwtSecret))
	}
	cfg.JWTSecret = jwtSecret

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
