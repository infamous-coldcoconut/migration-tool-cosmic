package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// možná přidat ještě rollback více stepů
func RunDown(dbUrl string, migrationsDir string) {
	db := ConnectDB(dbUrl)
	defer db.Close()

	AcquireLock(db)
	defer ReleaseLock(db)

	var lastVersion string
	err := db.QueryRow("SELECT version FROM cosmic_schema_migrations ORDER BY version DESC LIMIT 1").Scan(&lastVersion)
	
	if err == sql.ErrNoRows {
		fmt.Println("Database is clean, nothing to rollback.")
		return
	} else if err != nil {
		log.Fatalf("Error checking database state: %v", err)
	}

	fmt.Printf("Found latest migration to rollback: %s\n", lastVersion)

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var downFile string
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), lastVersion) && strings.HasSuffix(f.Name(), ".down.sql") {
			downFile = f.Name()
			break
		}
	}

	if downFile == "" {
		log.Fatalf("Critical error: Migration '%s' is in the database, but the corresponding .down.sql file is missing on the disk!", lastVersion)
	}

	// Spustí Rollback v izolované transakci
	filePath := filepath.Join(migrationsDir, downFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", downFile, err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec(string(content))
	if err != nil {
		tx.Rollback()
		log.Fatalf("Error executing %s: %v (ROLLBACK performed)", downFile, err)
	}

	_, err = tx.Exec("DELETE FROM cosmic_schema_migrations WHERE version = $1", lastVersion)
	if err != nil {
		tx.Rollback()
		log.Fatalf("Failed to delete migration state: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Printf("Successfully rolled back migration: %s\n", downFile)
}