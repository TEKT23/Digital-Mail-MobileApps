package main

import (
	"TugasAkhir/config"
	"TugasAkhir/models"
	"log"
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
