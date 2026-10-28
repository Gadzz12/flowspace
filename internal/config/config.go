package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi aplikasi dari .env dan environment
type Config struct {
	Port        string
	DBFile      string
	DatabaseURL string
}

// Load membaca konfigurasi dari file .env dan environment variables
func Load() Config {
	// Coba load .env memakai godotenv terlebih dahulu
	if err := godotenv.Load(); err != nil {
		// Fallback manual parser jika godotenv error atau file .env ada tapi godotenv tidak memprosesnya
		loadEnvManual(".env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		dbFile = "data.json"
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	return Config{
		Port:        port,
		DBFile:      dbFile,
		DatabaseURL: databaseURL,
	}
}

// loadEnvManual membaca file .env secara manual jika library mengalami kendala
func loadEnvManual(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}
