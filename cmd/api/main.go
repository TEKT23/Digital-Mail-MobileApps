package main

import (
	"TugasAkhir/config"
	"TugasAkhir/routes"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.ConnectDB()
	app := fiber.New()
		
	routes.Register(app)

	log.Println("🚀 API running on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
