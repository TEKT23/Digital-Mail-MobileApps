package routes

import (
	"TugasAkhir/handlers"
	"TugasAkhir/middleware"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/register", handlers.Register)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/forgot-password", handlers.RequestPasswordReset)
	auth.Post("/reset-password", handlers.ResetPassword)

	letters := api.Group("/letters",
		middleware.RequireAuth(),
		middleware.AuthorizeRoles(
			models.RoleBagianUmum,
			models.RoleADC,
			models.RoleDirektur,
			models.RoleAdmin,
		),
	)
	letters.Post("", handlers.CreateLetter)
	letters.Get("", handlers.ListLetters)
	letters.Get("/:id", handlers.GetLetterByID)
	letters.Put("/:id", handlers.UpdateLetter)
	letters.Delete("/:id", handlers.DeleteLetter)

	admin := api.Group("/admin",
		middleware.RequireAuth(),
		middleware.AuthorizeRoles(models.RoleAdmin),
	)
	admin.Post("/users", handlers.AdminCreateUser)
	admin.Get("/users", handlers.AdminListUsers) // ?page=&limit=&role=&q=
	admin.Get("/users/:id", handlers.AdminGetUserByID)
	admin.Put("/users/:id", handlers.AdminUpdateUser)
	admin.Delete("/users/:id", handlers.AdminDeleteUser)
}
