package main

import (
	"cloud-go-testtask/internal/config"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"log"
)

func main() {
	cfg := config.MustLoad()
	dsn := cfg.DBConfig.GetPostgresDSN()
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrationsDir := "./migrations"

	if err := goose.Up(db, migrationsDir); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
