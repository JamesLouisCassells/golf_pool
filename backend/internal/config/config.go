package config

import (
	"fmt"
	"os"
)

// Config holds runtime values loaded from the environment.
// Starting with a small struct now gives us a single place to grow config
// instead of spreading os.Getenv calls across the project.
type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	ClerkJWKSURL  string
	ClerkIssuer   string
	ClerkAudience string
	AdminClaim    string
	AdminValue    string
}

// Load reads configuration from environment variables and applies safe defaults
// for local development when a value has not been provided yet.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		ClerkJWKSURL:  os.Getenv("CLERK_JWKS_URL"),
		ClerkIssuer:   os.Getenv("CLERK_ISSUER"),
		ClerkAudience: os.Getenv("CLERK_AUDIENCE"),
		AdminClaim:    getEnv("CLERK_ADMIN_CLAIM", "role"),
		AdminValue:    getEnv("CLERK_ADMIN_VALUE", "admin"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
