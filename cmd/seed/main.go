package main

import (
	"log"

	"todo-app-test/internal/config"
	"todo-app-test/internal/database"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL belum diatur di file .env")
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Gagal AutoMigrate: %v", err)
	}

	if err := database.SeedFromJSON(db, cfg.DBFile); err != nil {
		log.Fatalf("Gagal seeding data dari %s: %v", cfg.DBFile, err)
	}

	log.Println("Seeding database selesai dengan sukses.")
}
