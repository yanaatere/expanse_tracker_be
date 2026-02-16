package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/yanaatere/expense_tracking/config"
)

func main() {
	// Initialize DB connection
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config and connect to DB:", err)
	}
	defer cfg.DB.Close()

	// migration folder path
	migrationsDir := "migrations"

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatal("Failed to read migrations directory:", err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sort to ensure execution order 001, 002, 003...
	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		fmt.Printf("Executing migration: %s... ", file)

		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("\nFailed to read file %s: %v", file, err)
		}

		// Execute SQL
		if _, err := cfg.DB.Exec(context.Background(), string(content)); err != nil {
			// Check if error is due to "already exists" to be deeper helpful,
			// but for now just fail or log.
			// Users usually want to know if it fails.
			log.Fatalf("\nFailed to execute migration %s: %v", file, err)
		}

		fmt.Println("Success!")
	}

	fmt.Println("Start Migration completed successfully.")
}
