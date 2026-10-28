package app

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"todo-app-test/internal/config"
	"todo-app-test/internal/database"
	"todo-app-test/internal/handler"
	"todo-app-test/internal/repository"
)

// Run menjalankan seluruh siklus hidup aplikasi Todo
func Run() {
	cfg := config.Load()

	var repo repository.Repository

	// Cek apakah DATABASE_URL tersedia di .env
	if cfg.DatabaseURL != "" {
		log.Println("Mencoba menghubungkan ke PostgreSQL melalui DATABASE_URL...")
		db, err := database.Open(cfg.DatabaseURL)
		if err != nil {
			log.Printf("PERINGATAN: Gagal koneksi ke database (%v). Menggunakan fallback penyimpanan %s.", err, cfg.DBFile)
			repo = repository.NewJSONRepository(cfg.DBFile)
		} else {
			if err := database.AutoMigrate(db); err != nil {
				log.Fatalf("Gagal melakukan AutoMigrate database: %v", err)
			}
			if err := database.SeedFromJSON(db, cfg.DBFile); err != nil {
				log.Printf("Peringatan seeding dari %s: %v", cfg.DBFile, err)
			}
			repo = repository.NewGORMRepository(db)
			log.Println("STATUS: Berhasil terhubung ke database PostgreSQL/Neon/Supabase.")
		}
	} else {
		log.Printf("STATUS: DATABASE_URL tidak diatur di .env. Menggunakan penyimpanan file %s.", cfg.DBFile)
		repo = repository.NewJSONRepository(cfg.DBFile)
	}

	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		// Fallback: coba dari direktori parent (jika dijalankan dari root proyek)
		if tmpl2, err2 := template.ParseGlob("../todo-list/templates/*.html"); err2 == nil {
			tmpl, err = tmpl2, nil
		} else {
			log.Fatal("Gagal membaca template HTML:", err)
		}
	}

	appHandler := handler.New(repo, tmpl)
	mux := http.NewServeMux()
	appHandler.RegisterRoutes(mux)

	fmt.Printf("Server berjalan di http://localhost%s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, mux))
}
