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

	// Users CRUD
	api.Post("/users", handlers.CreateUser)
	api.Get("/users", handlers.ListUsers)
	api.Get("/users/:id", handlers.GetUserByID)
	api.Put("/users/:id", handlers.UpdateUser)
	api.Delete("/users/:id", handlers.DeleteUser)
}
