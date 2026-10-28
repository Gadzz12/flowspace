package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"todo-app-test/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open membuka koneksi PostgreSQL memakai GORM dengan validasi DSN dan TLS
func Open(dsn string) (*gorm.DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("DATABASE_URL kosong")
	}

	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.User == nil || len(u.Path) < 2 {
		return nil, errors.New("DATABASE_URL wajib berupa connection string PostgreSQL Neon yang valid")
	}

	query := u.Query()
	switch query.Get("sslmode") {
	case "require", "verify-ca", "verify-full":
	default:
		return nil, errors.New("DATABASE_URL wajib memakai TLS: sslmode=require atau verify-full")
	}

	if query.Get("connect_timeout") == "" {
		query.Set("connect_timeout", "10")
	}
	u.RawQuery = query.Encode()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  u.String(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		TranslateError:      true,
		DisableAutomaticPing: true,
		Logger:              logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, errors.New("konfigurasi koneksi PostgreSQL tidak valid")
	}

	conn, err := db.DB()
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, errors.New("koneksi PostgreSQL gagal; periksa DATABASE_URL dan akses jaringan")
	}

	return db, nil
}

// AutoMigrate membuat/memperbarui tabel database secara otomatis
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Todo{},
		&model.SubTodo{},
		&model.Collaborator{},
		&model.Invitation{},
		&model.InboxMessage{},
	)
}

// SeedFromJSON menyalin data dari data.json ke database jika tabel masih kosong
func SeedFromJSON(db *gorm.DB, jsonPath string) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}

	// Jika tabel user sudah terisi, lewati seeding
	if count > 0 {
		return nil
	}

	file, err := os.ReadFile(jsonPath)
	if err != nil {
		// Jika file json tidak ditemukan, lewati
		return nil
	}

	var data model.DataStore
	if err := json.Unmarshal(file, &data); err != nil {
		return fmt.Errorf("format %s tidak valid: %w", jsonPath, err)
	}

	log.Println("Tabel database kosong. Memindahkan data dari", jsonPath, "ke database PostgreSQL...")

	return db.Transaction(func(tx *gorm.DB) error {
		for _, user := range data.Users {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		}

		for _, todo := range data.Todos {
			cleanTodo := todo
			cleanTodo.SubTodos = nil
			cleanTodo.Collaborators = nil
			cleanTodo.CollabRecords = nil
			if err := tx.Create(&cleanTodo).Error; err != nil {
				return err
			}
		}

		for _, sub := range data.SubTodos {
			if err := tx.Create(&sub).Error; err != nil {
				return err
			}
		}

		for _, collab := range data.Collaborators {
			if err := tx.Create(&collab).Error; err != nil {
				return err
			}
		}

		for _, inv := range data.Invitations {
			if err := tx.Create(&inv).Error; err != nil {
				return err
			}
		}

		for _, msg := range data.InboxMessages {
			if err := tx.Create(&msg).Error; err != nil {
				return err
			}
		}

		log.Println("Migrasi data dari JSON ke database berhasil dilakukan.")
		return nil
	})
}
