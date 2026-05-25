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
		log.Printf("Upozornění: MIGRATIONLOCKID '%s' není platné číslo, používám výchozí.", envLockID)
	}
	return 42424242
}

func ConnectDB(dbUrl string) *sql.DB {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Nelze inicializovat databázi: %v", err)
	}
	
	if err := db.Ping(); err != nil {
		log.Fatalf("Databáze neodpovídá. Zkontrolujte připojovací údaje a zda server běží. Detail: %v", err)
	}
	
	return db
}

func AcquireLock(db *sql.DB) {
	lockID := getMigrationLockID()
	var acquired bool

	err := db.QueryRow("SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	if err != nil {
		log.Fatalf("Chyba při snaze zamknout databázi: %v", err)
	}

	if !acquired {
		log.Fatalf("Databáze je aktuálně uzamčena jiným procesem! Zřejmě právě probíhá migrace na jiném serveru. Ukončuji...")
	}
	fmt.Println("🔒 Exkluzivní zámek databáze získán.")
}

// ReleaseLock zámek opět uvolní
func ReleaseLock(db *sql.DB) {
	lockID := getMigrationLockID()
	var released bool
	
	err := db.QueryRow("SELECT pg_advisory_unlock($1)", lockID).Scan(&released)
	if err != nil {
		log.Printf("Upozornění: Chyba při uvolňování zámku: %v", err)
	} else if released {
		fmt.Println("Zámek databáze úspěšně uvolněn.")
	}
}