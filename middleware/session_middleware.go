package middleware

import (
	"TugasAkhir/config"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Session store untuk admin web
var AdminSessionStore *session.Store

func init() {
	AdminSessionStore = session.New(session.Config{
		KeyLookup:      "cookie:admin_session",
		CookieSecure:   false, // Set true di production dengan HTTPS
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	})
}

const (
	SessionAdminIDKey   = "admin_id"
	SessionAdminRoleKey = "admin_role"
)

// RequireAdminSession - Middleware untuk cek session admin
func RequireAdminSession() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := AdminSessionStore.Get(c)
		if err != nil {
			return c.Redirect("/admin/login")
		}

		adminID := sess.Get(SessionAdminIDKey)
		adminRole := sess.Get(SessionAdminRoleKey)

		if adminID == nil || adminRole == nil {
			return c.Redirect("/admin/login")
		}

		// Cek apakah role adalah admin
		if adminRole != string(models.RoleAdmin) {
			sess.Destroy()
			return c.Redirect("/admin/login")
		}

		// Simpan di locals untuk digunakan di handler
		c.Locals(SessionAdminIDKey, adminID)
		c.Locals(SessionAdminRoleKey, adminRole)

		return c.Next()
	}
}

// GetAdminFromSession - Helper untuk mendapatkan admin user dari session
func GetAdminFromSession(c *fiber.Ctx) (*models.User, error) {
	adminID := c.Locals(SessionAdminIDKey)
	if adminID == nil {
		return nil, fiber.ErrUnauthorized
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", adminID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
