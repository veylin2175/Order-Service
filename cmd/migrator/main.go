package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	var (
		dbHost          string
		dbPort          int
		dbUser          string
		dbPassword      string
		dbName          string
		migrationPath   string
		migrationsTable string
	)

	// Парсим необходимые для миграций данные
	flag.StringVar(&dbHost, "db-host", "localhost", "database host")
	flag.IntVar(&dbPort, "db-port", 5432, "database port")
	flag.StringVar(&dbUser, "db-user", "", "database user")
	flag.StringVar(&dbPassword, "db-password", "", "database password")
	flag.StringVar(&dbName, "db-name", "", "database name")

	flag.StringVar(&migrationPath, "migration-path", "", "path to migration directory")
	flag.StringVar(&migrationsTable, "migrations-table", "", "migrations table name")
	flag.Parse()

	if dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("db-user, db-password, db-name are required")
	}
	if migrationPath == "" {
		log.Fatal("migration-path is required")
	}

	// Устанавливаем соединение с postgreSQL
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	// Создаем мигратор
	m, err := newMigrator(dsn, migrationPath, migrationsTable)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no migrations to apply")
			return
		}
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied successfully")
}

// newMigrator создает новый мигратор бд
func newMigrator(dsn, migrationPath, migrationTable string) (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		var err = db.Close()
		if err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: migrationTable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %v", err)
	}

	return m, nil
}
