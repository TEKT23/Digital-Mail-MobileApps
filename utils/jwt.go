package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"TugasAkhir/config"
	"TugasAkhir/models"
)

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type RegisteredClaims struct {
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
}

type JWTClaims struct {
	UserID    uint        `json:"user_id"`
	Role      models.Role `json:"role"`
	Email     string      `json:"email"`
	Username  string      `json:"username"`
	TokenType string      `json:"token_type,omitempty"`
	RegisteredClaims
}

func NewJWTClaimsFromUser(user models.User) JWTClaims {
	return JWTClaims{
		UserID:   user.ID,
		Role:     user.Role,
		Email:    user.Email,
		Username: user.Username,
	}
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

func signHS256(message string, secret []byte) string {
	return base64.RawURLEncoding.EncodeToString(signHS256Raw(message, secret))
}

func signHS256Raw(message string, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func generateToken(user models.User, tokenType string, ttl time.Duration) (string, *JWTClaims, error) {
	cfg := config.LoadJWTConfig()
	now := time.Now()

	claims := NewJWTClaimsFromUser(user)
	claims.TokenType = tokenType
	claims.RegisteredClaims = RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   strconv.FormatUint(uint64(user.ID), 10),
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	}

	headerJSON, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", nil, err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", nil, err
	}

	headerEnc := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsigned := headerEnc + "." + payloadEnc
	signature := signHS256(unsigned, cfg.SecretKey)

	token := unsigned + "." + signature
	return token, &claims, nil
}

func verifyToken(tokenString, expectedType string) (*JWTClaims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("token must have three parts")
	}

	cfg := config.LoadJWTConfig()

	unsigned := parts[0] + "." + parts[1]
	expectedSig := signHS256Raw(unsigned, cfg.SecretKey)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid token signature encoding")
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return nil, errors.New("invalid token signature")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token header encoding")
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, errors.New("invalid token header")
	}

	if header.Alg != "HS256" || header.Typ != "JWT" {
		return nil, errors.New("unsupported token header")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token payload encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, errors.New("invalid token payload")
	}

	if claims.Issuer != "" && claims.Issuer != cfg.Issuer {
		return nil, errors.New("invalid token issuer")
	}

	if claims.ExpiresAt != 0 && time.Now().Unix() >= claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}

	if expectedType != "" && !strings.EqualFold(claims.TokenType, expectedType) {
		return nil, errors.New("invalid token type")
	}

	return &claims, nil
}
