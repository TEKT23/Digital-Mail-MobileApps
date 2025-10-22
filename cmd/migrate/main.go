package main

import (
	config "TugasAkhir/config"
	"TugasAkhir/models"
	"log"
)

func main() {
	db := config.ConnectDB()

	if err := db.AutoMigrate(
		&models.User{},
		&models.Letter{},
	); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("✅ Migration success")
}
