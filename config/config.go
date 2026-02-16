package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DB *pgxpool.Pool
}

func LoadConfig() (*Config, error) {
	// Database configuration
	dbHost := getEnv("DB_HOST", "db.btcqmtnjujfkasfkffwo.supabase.co")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "natQkCKzJQnIcStx")
	dbName := getEnv("DB_NAME", "postgres")

	// Database connection string
	// Supabase requires SSL mode
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open database connection
	db, err := pgxpool.New(context.Background(), psqlInfo)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return &Config{
		DB: db,
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
