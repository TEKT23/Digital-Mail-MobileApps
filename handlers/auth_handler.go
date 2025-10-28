package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

type PasswordResetRequest struct {
	Email string `json:"email"`
}

func RequestPasswordReset(c *fiber.Ctx) error {
	var req PasswordResetRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
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

func generateResetToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(tokenBytes)
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:]), nil
}

func buildResetLink(token string) string {
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
