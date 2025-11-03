package main

import (
	"TugasAkhir/config"
	"TugasAkhir/models"
	"log"
)

func main() {
	db := config.ConnectDB()
	columnsToDrop := []string{"id", "created_at", "updated_at", "deleted_at"}
	for _, column := range columnsToDrop {
		if db.Migrator().HasColumn(&models.Letter{}, column) {
			if err := db.Migrator().DropColumn(&models.Letter{}, column); err != nil {
				log.Fatalf("failed to drop column %s: %v", column, err)
			}
			log.Printf("dropped obsolete column %s from surat table", column)
		}
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Letter{},
		&models.PasswordResetToken{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("✅ Migration completed")
}
