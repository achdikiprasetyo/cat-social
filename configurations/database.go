package configurations

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func DBConnection() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	params := os.Getenv("DB_PARAMS")
	// Buat string koneksi database
	dbConnectionString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?%s", dbUser, dbPassword, dbHost, dbPort, dbName, params)

	// Buat koneksi ke database
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Tes koneksi ke database
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("Connected to database")

	return db, nil
}
