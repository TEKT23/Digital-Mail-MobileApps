package routes

import (
	"TugasAkhir/handlers"
	"TugasAkhir/middleware"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	//Authentication
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/register", handlers.Register)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/forgot-password", handlers.RequestPasswordReset)
	auth.Post("/reset-password", handlers.ResetPassword)

	api.Use(middleware.RequireAuth())
	//Create Letter
	api.Post("/letters",
		middleware.AuthorizeRoles(
			models.RoleBagianUmum,
			models.RoleAdmin,
		),
		handlers.CreateLetter,
	)

	//Get All Letter
	api.Get("/letters",
		middleware.AuthorizeRoles(
			models.RoleBagianUmum,
			models.RoleADC,
			models.RoleDirektur,
			models.RoleAdmin,
		),
		handlers.ListLetters,
	)

	//Get Letter by ID
	api.Get("/letters/:id",
		middleware.AuthorizeRoles(
			models.RoleBagianUmum,
			models.RoleADC,
			models.RoleDirektur,
			models.RoleAdmin,
		),
		handlers.GetLetterByID,
	)

	//Update Letter
	api.Put("/letters/:id",
		middleware.AuthorizeRoles(
			models.RoleBagianUmum,
			models.RoleADC,
			models.RoleDirektur,
			models.RoleAdmin,
		),
		handlers.UpdateLetter,
	)

	//Delete Letter
	api.Delete("/letters/:id",
		middleware.AuthorizeRoles(
			models.RoleBagianUmum,
			models.RoleAdmin,
		),
		handlers.DeleteLetter,
	)

	//admin zone
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
