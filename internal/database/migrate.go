package database

import (
	"log"

	"siakad-mini/internal/models"
)

func Migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Student{},
		&models.Course{},
		&models.Enrollment{},
	)
	if err != nil {
		log.Fatalf("Migrasi gagal: %v", err)
	}
	log.Println("Migrasi berhasil")
}
