package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cosmic/utils"
)

func RunUp(dbUrl string, migrationsDir string) {
	db := ConnectDB(dbUrl)
	defer db.Close()

	AcquireLock(db)
	defer ReleaseLock(db)

	fmt.Println("Successfully connected to PostgreSQL database!")

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cosmic_schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		hash VARCHAR(255),
		author VARCHAR(255)
	)`)
	if err != nil {
		log.Fatalf("Failed to create cosmic_schema_migrations table: %v", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory (%s): %v", migrationsDir, err)
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	if len(upFiles) == 0 {
		fmt.Println("No migrations (.up.sql) found in the directory.")
		return
	}

	currentAuthor := utils.GetMigrationAuthor()
	appliedCount := 0

	for _, file := range upFiles {
		version := strings.TrimSuffix(file, ".up.sql")

		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", file, err)
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			log.Fatalf("Error: File %s is empty! Please write and save your SQL commands before running 'up'.", file)
		}

		currentHash := utils.CalculateHash(content)

		var dbHash string 
		err = db.QueryRow("SELECT hash FROM cosmic_schema_migrations WHERE version = $1", version).Scan(&dbHash)
		
		if err == nil {
			// Záznam v DB existuje (migrace už proběhla)
			if dbHash != currentHash {
				log.Fatalf(`
CRITICAL INTEGRITY ERROR! 
File: %s

This migration was successfully applied in the past, but its content on disk has CHANGED!
It is strictly forbidden to modify historical SQL files. If you need a change, create a new migration.

Expected hash (in DB): %s
Actual hash (file): %s
`, file, dbHash, currentHash)
			}
			
			// Hashe sedí, přeskočí
			continue
		} else if err != sql.ErrNoRows {
			log.Fatalf("Error checking status for migration %s: %v", version, err)
		}

		// Migrace ještě neproběhla -> Transakce
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction for %s: %v", file, err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback() 
			log.Fatalf("Error executing %s: %v (ROLLBACK performed)", file, err)
		}

		_, err = tx.Exec("INSERT INTO cosmic_schema_migrations (version, hash, author) VALUES ($1, $2, $3)", version, currentHash, currentAuthor)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to save migration state %s: %v", file, err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatalf("Failed to commit transaction for %s: %v", file, err)
		}

		fmt.Printf("Applied migration: %s (Author: %s)\n", file, currentAuthor)
		appliedCount++
	}

	if appliedCount == 0 {
		fmt.Println("Database is up-to-date, no new migrations to apply.")
	} else {
		fmt.Printf("Successfully applied %d new migrations!\n", appliedCount)
	}
}