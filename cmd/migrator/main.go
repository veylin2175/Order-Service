package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Config struct {
	DBHost          string
	DBPort          int
	DBUser          string
	DBPass          string
	DBName          string
	MigrationPath   string
	MigrationsTable string
}

func main() {
	cfg := parseFlags()

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		var err = db.Close()
		if err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}(db)

	if err := runMigrations(db, cfg); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("Migrations completed successfully")
}

func parseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.DBHost, "db-host", "localhost", "Database host")
	flag.IntVar(&cfg.DBPort, "db-port", 5432, "Database port")
	flag.StringVar(&cfg.DBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.DBPass, "db-pass", "", "Database password")
	flag.StringVar(&cfg.DBName, "db-name", "", "Database name")
	flag.StringVar(&cfg.MigrationPath, "migration-path", "./migrations", "Path to migrations directory")
	flag.StringVar(&cfg.MigrationsTable, "migrations-table", "goose_db_version", "Migrations table name")

	flag.Parse()

	if cfg.DBUser == "" || cfg.DBPass == "" || cfg.DBName == "" {
		log.Fatal("db-user, db-pass and db-name are required parameters")
	}

	return cfg
}

func runMigrations(db *sql.DB, cfg Config) error {
	// Получаем абсолютный путь к миграциям
	absPath, err := filepath.Abs(cfg.MigrationPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Проверка существования директории
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist at: %s", absPath)
	}

	log.Printf("Using migrations from: %s", absPath)

	// Настройка goose с абсолютным путем
	goose.SetBaseFS(os.DirFS(absPath))
	goose.SetTableName(cfg.MigrationsTable)
	goose.SetSequential(true)

	// Применение миграций
	if err := goose.Up(db, absPath); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
