package config

import "os"

// Config holds runtime values loaded from the environment.
// Starting with a small struct now gives us a single place to grow config
// instead of spreading os.Getenv calls across the project.
type Config struct {
	HTTPAddr string
}

// Load reads configuration from environment variables and applies safe defaults
// for local development when a value has not been provided yet.
func Load() Config {
	return Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
