package handlers

import (
	"TugasAkhir/dto"
	"TugasAkhir/utils"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"

	"TugasAkhir/config"
	"TugasAkhir/models"
	"TugasAkhir/utils/mailer"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "email and password are required", nil)
	}

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.JSONError(c, fiber.StatusUnauthorized, "invalid email or password", nil)
		}
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to fetch user", err.Error())
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return utils.JSONError(c, fiber.StatusUnauthorized, "invalid email or password", nil)
	}

	accessToken, accessClaims, err := utils.GenerateAccessToken(user)
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to generate access token", err.Error())
	}

	refreshToken, _, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to generate refresh token", err.Error())
	}

	resp := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    time.Unix(accessClaims.ExpiresAt, 0),
		User:         toUserSummary(user),
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "login successful", resp)
}

func Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	req.Username = strings.TrimSpace(req.Username)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.Jabatan = strings.TrimSpace(req.Jabatan)
	req.Atribut = strings.TrimSpace(req.Atribut)

	if req.Username == "" || len(req.Username) < 3 {
		return utils.JSONError(c, fiber.StatusBadRequest, "username must be at least 3 characters", nil)
	}
	if req.Email == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "email is required", nil)
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid email format", nil)
	}
	if req.Password == "" || len(req.Password) < 8 {
		return utils.JSONError(c, fiber.StatusBadRequest, "password must be at least 8 characters", nil)
	}
	if !isValidRole(req.Role) {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid role provided", nil)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to process password", err.Error())
	}

	user := models.User{
		Username:     req.Username,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		Jabatan:      req.Jabatan,
		Atribut:      req.Atribut,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		if isDuplicateEntryError(err) {
			return utils.JSONError(c, fiber.StatusBadRequest, "username or email already exists", nil)
		}
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to create user", err.Error())
	}

	resp := dto.RegisterResponse{
		User:    toUserSummary(user),
		Message: "registration successful",
	}

	return utils.JSONSuccess(c, fiber.StatusCreated, "registration successful", resp)
}

func RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	token := strings.TrimSpace(req.RefreshToken)
	if token == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "refresh token is required", nil)
	}

	claims, err := utils.VerifyRefreshToken(token)
	if err != nil {
		return utils.JSONError(c, fiber.StatusUnauthorized, "invalid or expired refresh token", nil)
	}

	var user models.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.JSONError(c, fiber.StatusUnauthorized, "user no longer exists", nil)
		}
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to fetch user", err.Error())
	}

	accessToken, accessClaims, err := utils.GenerateAccessToken(user)
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to generate access token", err.Error())
	}

	refreshToken, _, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to generate refresh token", err.Error())
	}

	resp := dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    time.Unix(accessClaims.ExpiresAt, 0),
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "token refreshed successfully", resp)
}

func RequestPasswordReset(c *fiber.Ctx) error {
	var req dto.PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "email is required", nil)
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid email format", nil)
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.JSONSuccess(c, fiber.StatusOK, "if the email exists, a reset link has been sent", nil)
		}
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to fetch user", err.Error())
	}

	usedAt := time.Now()
	if err := config.DB.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used = ?", user.ID, false).
		Updates(map[string]any{
			"used":    true,
			"used_at": &usedAt,
		}).Error; err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to invalidate previous reset tokens", err.Error())
	}

	rawToken, tokenHash, err := generateResetToken()
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to generate token", err.Error())
	}

	resetToken := models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := config.DB.Create(&resetToken).Error; err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to store reset token", err.Error())
	}

	resetLink := buildResetLink(rawToken)
	emailCfg := config.LoadEmailConfig()
	mailClient := mailer.NewClient(emailCfg)
	if err := mailClient.SendPasswordResetEmail(user.Email, resetLink); err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to send reset email", err.Error())
	}
	return utils.JSONSuccess(c, fiber.StatusOK, "if the email exists, a reset link has been sent", nil)
}

func ResetPassword(c *fiber.Ctx) error {
	var req dto.PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	req.Token = strings.TrimSpace(req.Token)
	req.Password = strings.TrimSpace(req.Password)

	if req.Token == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "token is required", nil)
	}
	if len(req.Password) < 8 {
		return utils.JSONError(c, fiber.StatusBadRequest, "password must be at least 8 characters", nil)
	}
	if req.Password != req.ConfirmPassword {
		return utils.JSONError(c, fiber.StatusBadRequest, "password confirmation does not match", nil)
	}

	tokenHash := hashToken(req.Token)

	var reset models.PasswordResetToken
	if err := config.DB.Preload("User").Where("token_hash = ?", tokenHash).First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.JSONError(c, fiber.StatusBadRequest, "invalid or expired token", nil)
		}
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to fetch reset token", err.Error())
	}

	if reset.Used || time.Now().After(reset.ExpiresAt) {
		return utils.JSONError(c, fiber.StatusBadRequest, "invalid or expired token", nil)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to process password", err.Error())
	}

	now := time.Now()
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).
			Where("id = ?", reset.UserID).
			Update("password_hash", hashedPassword).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.PasswordResetToken{}).
			Where("id = ?", reset.ID).
			Updates(map[string]any{
				"used":    true,
				"used_at": &now,
			}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return utils.JSONError(c, fiber.StatusInternalServerError, "failed to reset password", err.Error())
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "password has been reset successfully", nil)
}

func toUserSummary(user models.User) dto.UserSummary {
	return dto.UserSummary{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		Jabatan:   user.Jabatan,
		Atribut:   user.Atribut,
	}
}

func isValidRole(role models.Role) bool {
	switch role {
	case models.RoleBagianUmum, models.RoleADC, models.RoleDirektur, models.RoleAdmin:
		return true
	default:
		return false
	}
}

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "unique constraint")
}

func generateResetToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(tokenBytes)
	return raw, hashToken(raw), nil
}

func buildResetLink(token string) string {
	config.LoadEnv()
	base := os.Getenv("PASSWORD_RESET_URL")
	if base == "" {
		base = "/auth/reset-password"
	}

	escapedToken := url.QueryEscape(token)
	if strings.Contains(base, "?") {
		if strings.HasSuffix(base, "?") || strings.HasSuffix(base, "&") {
			return base + "token=" + escapedToken
		}
		return base + "&token=" + escapedToken
	}
	return base + "?token=" + escapedToken
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
