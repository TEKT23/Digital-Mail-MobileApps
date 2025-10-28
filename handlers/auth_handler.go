package handlers

import (
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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetSubmission struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func RequestPasswordReset(c *fiber.Ctx) error {
	var req PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid email format"})
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	usedAt := time.Now()
	if err := config.DB.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used = ?", user.ID, false).
		Updates(map[string]any{
			"used":    true,
			"used_at": &usedAt,
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to invalidate previous reset tokens"})
	}

	rawToken, tokenHash, err := generateResetToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	resetToken := models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := config.DB.Create(&resetToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	resetLink := buildResetLink(rawToken)
	emailCfg := config.LoadEmailConfig()
	mailClient := mailer.NewClient(emailCfg)
	if err := mailClient.SendPasswordResetEmail(user.Email, resetLink); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
}

func ResetPassword(c *fiber.Ctx) error {
	var req PasswordResetSubmission
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	req.Token = strings.TrimSpace(req.Token)
	req.Password = strings.TrimSpace(req.Password)

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
	}
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 8 characters"})
	}

	tokenHash := hashToken(req.Token)

	var reset models.PasswordResetToken
	if err := config.DB.Preload("User").Where("token_hash = ?", tokenHash).First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired token"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if reset.Used || time.Now().After(reset.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired token"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process password"})
	}

	now := time.Now()
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).
			Where("id = ?", reset.UserID).
			Update("password_hash", string(hashedPassword)).Error; err != nil {
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Password has been reset successfully"})
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
