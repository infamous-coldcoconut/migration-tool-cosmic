package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
)

func getMigrationLockID() int64 {
	envLockID := os.Getenv("MIGRATIONLOCKID")
	if envLockID != "" {
		// Převedení textu na 64-bitové číslo (Base 10)
		parsedID, err := strconv.ParseInt(envLockID, 10, 64)
		if err == nil {
			return parsedID
		}
		log.Printf("Warning: MIGRATIONLOCKID '%s' is not a valid number, using default.", envLockID)
	}
	return 42424242
}

func ConnectDB(dbUrl string) *sql.DB {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unresponsive. Check connection details and whether the server is running. Details: %v", err)
	}
	
	return db
}

func AcquireLock(db *sql.DB) {
	lockID := getMigrationLockID()
	var acquired bool

	err := db.QueryRow("SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	if err != nil {
		log.Fatalf("Error while trying to lock the database: %v", err)
	}

	if !acquired {
		log.Fatalf("Database is currently locked by another process! A migration is likely running on another server. Exiting...")
	}
	fmt.Println("Exclusive database lock acquired.")
}

// ReleaseLock zámek opět uvolní
func ReleaseLock(db *sql.DB) {
	lockID := getMigrationLockID()
	var released bool
	
	err := db.QueryRow("SELECT pg_advisory_unlock($1)", lockID).Scan(&released)
	if err != nil {
		log.Printf("Warning: Error while releasing lock: %v", err)
	} else if released {
		fmt.Println("Database lock successfully released.")
	}
}