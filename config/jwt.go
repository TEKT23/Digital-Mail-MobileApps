package config

import (
	"log"
	"os"
	"sync"
	"time"
)

type JWTConfig struct {
	SecretKey       []byte
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

var (
	jwtConfig JWTConfig
	jwtOnce   sync.Once
)

func LoadJWTConfig() JWTConfig {
	jwtOnce.Do(func() {
		LoadEnv()

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Fatal("JWT_SECRET environment variable is not set")
		}

		issuer := os.Getenv("JWT_ISSUER")
		if issuer == "" {
			issuer = "digital-mail"
		}

		ttl := 24 * time.Hour
		if ttlStr := os.Getenv("JWT_ACCESS_TTL"); ttlStr != "" {
			if parsed, err := time.ParseDuration(ttlStr); err == nil {
				ttl = parsed
			} else {
				log.Printf("invalid JWT_ACCESS_TTL value %q, using default %s", ttlStr, ttl)
			}
		}

		refreshTTL := 7 * 24 * time.Hour
		if ttlStr := os.Getenv("JWT_REFRESH_TTL"); ttlStr != "" {
			if parsed, err := time.ParseDuration(ttlStr); err == nil {
				refreshTTL = parsed
			} else {
				log.Printf("invalid JWT_REFRESH_TTL value %q, using default %s", ttlStr, refreshTTL)
			}
		}

		jwtConfig = JWTConfig{
			SecretKey:       []byte(secret),
			Issuer:          issuer,
			AccessTokenTTL:  ttl,
			RefreshTokenTTL: ttl,
		}
	})

	return jwtConfig
}
