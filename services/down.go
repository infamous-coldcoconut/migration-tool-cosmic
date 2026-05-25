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
		fmt.Println("Databáze je čistá, není co vracet (rollback).")
		return
	} else if err != nil {
		log.Fatalf("Chyba při zjišťování stavu databáze: %v", err)
	}

	fmt.Printf("Zjištěna poslední migrace k vrácení: %s\n", lastVersion)

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Nelze načíst složku s migracemi: %v", err)
	}

	var downFile string
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), lastVersion) && strings.HasSuffix(f.Name(), ".down.sql") {
			downFile = f.Name()
			break
		}
	}

	if downFile == "" {
		log.Fatalf("Kritická chyba: V databázi je migrace '%s', ale na disku chybí odpovídající .down.sql soubor pro rollback!", lastVersion)
	}

	// Spustí Rollback v izolované transakci
	filePath := filepath.Join(migrationsDir, downFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Nelze přečíst soubor %s: %v", downFile, err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Nelze zahájit transakci: %v", err)
	}

	_, err = tx.Exec(string(content))
	if err != nil {
		tx.Rollback()
		log.Fatalf("Chyba při exekuci %s: %v (proveden ROLLBACK)", downFile, err)
	}

	_, err = tx.Exec("DELETE FROM cosmic_schema_migrations WHERE version = $1", lastVersion)
	if err != nil {
		tx.Rollback()
		log.Fatalf("Nelze smazat stav migrace: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Nelze potvrdit transakci: %v", err)
	}

	fmt.Printf("Úspěšně vrácena migrace: %s\n", downFile)
}