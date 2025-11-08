package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"TugasAkhir/config"
	"TugasAkhir/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID    uint        `json:"user_id"`
	Role      models.Role `json:"role"`
	Email     string      `json:"email"`
	Username  string      `json:"username"`
	TokenType string      `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(user models.User) (string, *JWTClaims, error) {
	cfg := config.LoadJWTConfig()
	return generateToken(user, "access", cfg.AccessTokenTTL)
}

func VerifyAccessToken(tokenString string) (*JWTClaims, error) {
	return verifyToken(tokenString, "access")
}

func GenerateRefreshToken(user models.User) (string, *JWTClaims, error) {
	cfg := config.LoadJWTConfig()
	return generateToken(user, "refresh", cfg.RefreshTokenTTL)
}

func VerifyRefreshToken(tokenString string) (*JWTClaims, error) {
	return verifyToken(tokenString, "refresh")
}

func generateToken(user models.User, tokenType string, ttl time.Duration) (string, *JWTClaims, error) {
	cfg := config.LoadJWTConfig()
	now := time.Now()

	claims := &JWTClaims{
		UserID:    user.ID,
		Role:      user.Role,
		Email:     user.Email,
		Username:  user.Username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(cfg.SecretKey)
	if err != nil {
		return "", nil, err
	}
	return signed, claims, nil
}

func verifyToken(tokenString, expectedType string) (*JWTClaims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}

	cfg := config.LoadJWTConfig()

	claims := &JWTClaims{}
	parsed, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return cfg.SecretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(cfg.Issuer),
	)
	if err != nil {
		return nil, err
	}

	if !parsed.Valid {
		return nil, errors.New("invalid token ")
	}

	if expectedType != "" && !strings.EqualFold(claims.TokenType, expectedType) {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}
