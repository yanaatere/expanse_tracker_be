package config

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

const MinioBucket = "receipts"

type Config struct {
	DB             *pgxpool.Pool
	Redis          *redis.Client
	Minio          *minio.Client
	MinioPublicURL string
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

	log.Printf("Config loaded - Host: %s, Port: %s, DB: %s", dbHost, dbPort, dbName)

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

	// Redis connection
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis unavailable (%v) — bot linking will not work", err)
	}

	// MinIO configuration
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccess := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecret := getEnv("MINIO_SECRET_KEY", "minioadmin")
	minioUseSSL := getEnv("MINIO_USE_SSL", "false") == "true"
	minioPublicURL := getEnv("MINIO_PUBLIC_URL", "http://localhost:9000")

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccess, minioSecret, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Ensure bucket exists with public read policy
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check minio bucket: %w", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create minio bucket: %w", err)
		}
		log.Printf("Created MinIO bucket: %s", MinioBucket)
	}

	// Bucket is private — receipts are served via pre-signed URLs (see upload_handler.go).

	return &Config{
		DB:             db,
		Redis:          redisClient,
		Minio:          minioClient,
		MinioPublicURL: minioPublicURL,
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
