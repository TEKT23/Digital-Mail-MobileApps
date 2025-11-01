package middleware

import (
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
)

func AuthorizeRoles(allowedRoles ...models.Role) fiber.Handler {
	allowed := make(map[models.Role]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		claims, ok := GetJWTClaims(c)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "authorization context missing"})
		}

		if len(allowed) == 0 {
			return c.Next()
		}

		if _, ok := allowed[claims.Role]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
		}

		return c.Next()
	}
}
