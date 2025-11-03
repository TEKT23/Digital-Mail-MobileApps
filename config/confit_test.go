package config

import "testing"

func TestValidateDatabaseConfigMissing(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASS", "")
	t.Setenv("DB_NAME", "")

	if err := ValidateDatabaseConfig(); err == nil {
		t.Fatal("expected validation error for missing database environment variables")
	}
}

func TestValidateDatabaseConfigSuccess(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASS", "pass")
	t.Setenv("DB_NAME", "dbname")

	if err := ValidateDatabaseConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateJWTConfigMissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	if err := ValidateJWTConfig(); err == nil {
		t.Fatal("expected validation error for missing JWT secret")
	}
}

func TestValidateJWTConfigInvalidTTL(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_ACCESS_TTL", "not-a-duration")

	if err := ValidateJWTConfig(); err == nil {
		t.Fatal("expected validation error for invalid JWT access TTL")
	}
}

func TestValidateEmailConfigInvalidPort(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "invalid")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "from@example.com")

	if err := ValidateEmailConfig(); err == nil {
		t.Fatal("expected validation error for invalid SMTP_PORT")
	}
}

func TestValidateEmailConfigSuccess(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "from@example.com")

	if err := ValidateEmailConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAggregatesSections(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASS", "pass")
	t.Setenv("DB_NAME", "dbname")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "from@example.com")

	if err := Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
