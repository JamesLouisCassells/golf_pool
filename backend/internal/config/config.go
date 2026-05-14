package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds runtime values loaded from the environment.
// Starting with a small struct now gives us a single place to grow config
// instead of spreading os.Getenv calls across the project.
type Config struct {
	HTTPAddr               string
	DatabaseURL            string
	MockAuthEnabled        bool
	MockAuthClerkID        string
	MockAuthEmail          string
	MockAuthName           string
	MockAuthAdmin          bool
	ClerkJWKSURL           string
	ClerkIssuer            string
	ClerkAudience          string
	ClerkSecretKey         string
	ClerkAuthorizedParties []string
	ClerkEmailClaim        string
	ClerkNameClaim         string
	AdminClaim             string
	AdminValue             string
	GolfProvider           string
	GolfAPIBaseURL         string
	GolfAPIKey             string
	GolfAPIHost            string
}

// Load reads configuration from environment variables and applies safe defaults
// for local development when a value has not been provided yet.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:               getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		MockAuthEnabled:        getEnvBool("MOCK_AUTH_ENABLED", false),
		MockAuthClerkID:        getEnv("MOCK_AUTH_CLERK_ID", "dev-user"),
		MockAuthEmail:          getEnv("MOCK_AUTH_EMAIL", "dev@example.com"),
		MockAuthName:           getEnv("MOCK_AUTH_NAME", "Dev User"),
		MockAuthAdmin:          getEnvBool("MOCK_AUTH_ADMIN", false),
		ClerkJWKSURL:           os.Getenv("CLERK_JWKS_URL"),
		ClerkIssuer:            os.Getenv("CLERK_ISSUER"),
		ClerkAudience:          os.Getenv("CLERK_AUDIENCE"),
		ClerkSecretKey:         os.Getenv("CLERK_SECRET_KEY"),
		ClerkAuthorizedParties: splitCSV(os.Getenv("CLERK_AUTHORIZED_PARTIES")),
		ClerkEmailClaim:        getEnv("CLERK_EMAIL_CLAIM", "email"),
		ClerkNameClaim:         getEnv("CLERK_NAME_CLAIM", "name"),
		AdminClaim:             getEnv("CLERK_ADMIN_CLAIM", "role"),
		AdminValue:             getEnv("CLERK_ADMIN_VALUE", "admin"),
		GolfProvider:           strings.TrimSpace(strings.ToLower(os.Getenv("GOLF_PROVIDER"))),
		GolfAPIBaseURL:         strings.TrimSpace(os.Getenv("GOLF_API_BASE_URL")),
		GolfAPIKey:             strings.TrimSpace(os.Getenv("GOLF_API_KEY")),
		GolfAPIHost:            strings.TrimSpace(os.Getenv("GOLF_API_HOST")),
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

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
