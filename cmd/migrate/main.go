package main

import (
	"TugasAkhir/config"
	"TugasAkhir/models"
	"TugasAkhir/routes"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db := config.ConnectDB()
	if err := db.AutoMigrate(
		&models.User{},
		&models.Letter{},
		&models.PasswordResetToken{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("✅ Migration completed")
}
