package routes

import (
	"TugasAkhir/handlers"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	// CRUD Surat
	api.Post("/letters", handlers.CreateLetter)
	api.Get("/letters", handlers.ListLetters)
	api.Get("/letters/:id", handlers.GetLetterByID)
	api.Put("/letters/:id", handlers.UpdateLetter)
	api.Delete("/letters/:id", handlers.DeleteLetter)
}
