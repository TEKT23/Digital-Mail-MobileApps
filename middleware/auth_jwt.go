package middleware

import (
	"strings"

	"TugasAkhir/utils"

	"github.com/gofiber/fiber/v2"
)

const (
	ContextClaimsKey   = "jwtClaims"
	ContextUserIDKey   = "userID"
	ContextUserRoleKey = "userRole"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing Authorization header"})
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid Authorization header"})
		}

		tokenString := strings.TrimSpace(parts[1])
		claims, err := utils.VerifyAccessToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		c.Locals(ContextClaimsKey, claims)
		c.Locals(ContextUserIDKey, claims.UserID)
		c.Locals(ContextUserRoleKey, claims.Role)

		return c.Next()
	}
}

func GetJWTClaims(c *fiber.Ctx) (*utils.JWTClaims, bool) {
	claims, ok := c.Locals(ContextClaimsKey).(*utils.JWTClaims)
	return claims, ok
}
