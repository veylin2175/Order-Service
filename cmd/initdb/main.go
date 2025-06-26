package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"L0/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	// Загрузка конфигурации
	cfg := config.MustLoad()

	// Пароль postgres берем из переменной окружения
	adminPassword := os.Getenv("POSTGRES_ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("POSTGRES_ADMIN_PASSWORD environment variable not set")
	}

	// DSN для подключения с правами администратора
	adminDSN := fmt.Sprintf("host=%s port=%d user=postgres password=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		adminPassword,
		cfg.Database.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Подключение к PostgreSQL
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Выполнение SQL скрипта инициализации
	if err := initDatabase(db, cfg); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	log.Println("Database initialized successfully")
}

func initDatabase(db *sql.DB, cfg *config.Config) error {
	// Проверка существования БД
	var dbExists bool
	err := db.QueryRowContext(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.Database.DBName,
	).Scan(&dbExists)

	if err != nil {
		return fmt.Errorf("database check failed: %w", err)
	}

	if !dbExists {
		if _, err := db.Exec(fmt.Sprintf(
			"CREATE DATABASE %s WITH ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8' TEMPLATE template0",
			cfg.Database.DBName,
		)); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database %s created", cfg.Database.DBName)
	}

	initScript := `
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = '%s') THEN
        CREATE USER %s WITH PASSWORD '%s';
    END IF;
END
$$;

GRANT CONNECT ON DATABASE %s TO %s;

GRANT USAGE, CREATE ON SCHEMA public TO %s;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO %s;
`

	// Форматируем скрипт с параметрами из конфига
	script := fmt.Sprintf(initScript,
		cfg.Database.User, cfg.Database.User, cfg.Database.Password, // Пользователь
		cfg.Database.DBName, cfg.Database.User, // Права на БД
		cfg.Database.User,                    // Права на схему
		cfg.Database.User, cfg.Database.User, // Права на таблицы и последовательности
	)

	_, err = db.Exec(script)
	return err
}
