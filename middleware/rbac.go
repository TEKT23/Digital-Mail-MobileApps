package middleware

import (
	"TugasAkhir/models"
	"TugasAkhir/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RequireRole(allowedRoles ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("jwtClaims").(*utils.JWTClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		for _, role := range allowedRoles {
			if claims.Role == role {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}
}

func RequireStaf() fiber.Handler {
	return RequireRole(models.RoleStafProgram, models.RoleStafLembaga)
}
func RequireManajer() fiber.Handler {
	return RequireRole(models.RoleManajerKPP, models.RoleManajerPemas, models.RoleManajerPKL)
}
func RequireDirektur() fiber.Handler { return RequireRole(models.RoleDirektur) }
func RequireAdmin() fiber.Handler    { return RequireRole(models.RoleAdmin) }

func GetUserFromContext(c *fiber.Ctx) (*models.User, error) {
	claims, ok := c.Locals("jwtClaims").(*utils.JWTClaims)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	return &models.User{
		Model: gorm.Model{ID: claims.UserID},
		Role:  claims.Role,
		Email: claims.Email,
	}, nil
}
