package config

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	DB *pgxpool.Pool
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "postgres")

	// Log the configuration for debugging
	log.Printf("Config loaded - Host: %s, Port: %s, User: %s, DB: %s", dbHost, dbPort, dbUser, dbName)

	// Determine SSL mode based on host
	// Use SSL only for remote databases (not localhost or docker service name)
	sslMode := "disable"
	isRemoteHost := dbHost != "localhost" && dbHost != "db" && dbHost != "127.0.0.1"
	if isRemoteHost {
		sslMode = "require"
	}

	// Database connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)

	// Parse the connection config
	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	// Force IPv4 for remote hosts (Docker containers often lack IPv6 connectivity)
	if isRemoteHost {
		poolConfig.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "tcp4", addr)
		}
	}

	// Open database connection
	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
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
