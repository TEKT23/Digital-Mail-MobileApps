package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Validate ensures all configuration sections have the required environment
// variables set and that optional values are well-formed.
func Validate() error {
	LoadEnv()

	if err := ValidateDatabaseConfig(); err != nil {
		return fmt.Errorf("database configuration: %w", err)
	}

	if err := ValidateJWTConfig(); err != nil {
		return fmt.Errorf("jwt configuration: %w", err)
	}

	if err := ValidateEmailConfig(); err != nil {
		return fmt.Errorf("email configuration: %w", err)
	}

	return nil
}

// ValidateDatabaseConfig ensures all required database environment variables
// are present.
func ValidateDatabaseConfig() error {
	required := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASS", "DB_NAME"}

	var missing []string
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateJWTConfig ensures JWT environment variables are set and valid.
func ValidateJWTConfig() error {
	if strings.TrimSpace(os.Getenv("JWT_SECRET")) == "" {
		return fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	if ttl := strings.TrimSpace(os.Getenv("JWT_ACCESS_TTL")); ttl != "" {
		if _, err := time.ParseDuration(ttl); err != nil {
			return fmt.Errorf("invalid JWT_ACCESS_TTL value %q: %w", ttl, err)
		}
	}

	if ttl := strings.TrimSpace(os.Getenv("JWT_REFRESH_TTL")); ttl != "" {
		if _, err := time.ParseDuration(ttl); err != nil {
			return fmt.Errorf("invalid JWT_REFRESH_TTL value %q: %w", ttl, err)
		}
	}

	return nil
}

// ValidateEmailConfig ensures email configuration values are provided and valid.
func ValidateEmailConfig() error {
	required := []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM"}

	var missing []string
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil || port <= 0 {
		return fmt.Errorf("SMTP_PORT must be a positive integer")
	}

	return nil
}
