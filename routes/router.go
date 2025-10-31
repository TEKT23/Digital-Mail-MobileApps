package routes

import (
	"TugasAkhir/handlers"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	// Letters CRUD
	api.Post("/letters", handlers.CreateLetter)
	api.Get("/letters", handlers.ListLetters)
	api.Get("/letters/:id", handlers.GetLetterByID)
	api.Put("/letters/:id", handlers.UpdateLetter)
	api.Delete("/letters/:id", handlers.DeleteLetter)

	// ----- ADMIN USERS CRUD -----
	admin := api.Group("/admin")
	// TODO: pasang middleware admin-only di sini (JWT + role check)
	admin.Post("/users", handlers.AdminCreateUser)
	admin.Get("/users", handlers.AdminListUsers) // ?page=&limit=&role=&q=
	admin.Get("/users/:id", handlers.AdminGetUserByID)
	admin.Put("/users/:id", handlers.AdminUpdateUser)
	admin.Delete("/users/:id", handlers.AdminDeleteUser)

	// Auth
	api.Post("/auth/forgot-password", handlers.RequestPasswordReset)
	api.Post("/auth/reset-password", handlers.ResetPassword)
}
