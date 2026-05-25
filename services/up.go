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

	fmt.Println("Úspěšně připojeno k PostgreSQL databázi!")

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cosmic_schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		hash VARCHAR(255),
		author VARCHAR(255)
	)`)
	if err != nil {
		log.Fatalf("Nelze vytvořit tabulku cosmic_schema_migrations: %v", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Nelze načíst složku s migracemi (%s): %v", migrationsDir, err)
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	if len(upFiles) == 0 {
		fmt.Println("Ve složce nejsou žádné migrace (.up.sql).")
		return
	}

	currentAuthor := utils.GetMigrationAuthor()
	appliedCount := 0

	for _, file := range upFiles {
		version := strings.TrimSuffix(file, ".up.sql")

		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Nelze přečíst soubor %s: %v", file, err)
		}

		currentHash := utils.CalculateHash(content)

		var dbHash string 
		err = db.QueryRow("SELECT hash FROM cosmic_schema_migrations WHERE version = $1", version).Scan(&dbHash)
		
		if err == nil {
			// Záznam v DB existuje (migrace už proběhla)
			if dbHash != currentHash {
				log.Fatalf(`
				KRITICKÁ CHYBA INTEGRITY! 
				Soubor: %s

				Tato migrace už byla v minulosti úspěšně spuštěna, ale její obsah na disku se ZMĚNIL!
				Očekávaný otisk (v DB): %s
				Aktuální otisk (soubor): %s
				`, file, dbHash, currentHash)
			}
			
			// Hashe sedí, přeskočí
			continue
		} else if err != sql.ErrNoRows {
			log.Fatalf("Chyba při kontrole stavu migrace %s: %v", version, err)
		}

		// Migrace ještě neproběhla -> Transakce
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Nelze zahájit transakci pro %s: %v", file, err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback() 
			log.Fatalf("Chyba při exekuci %s: %v (proveden ROLLBACK)", file, err)
		}

		_, err = tx.Exec("INSERT INTO cosmic_schema_migrations (version, hash, author) VALUES ($1, $2, $3)", version, currentHash, currentAuthor)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Nelze zapsat stav migrace %s: %v", file, err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatalf("Nelze potvrdit transakci pro %s: %v", file, err)
		}

		fmt.Printf("Aplikována migrace: %s (Autor: %s)\n", file, currentAuthor)
		appliedCount++
	}

	if appliedCount == 0 {
		fmt.Println("Databáze je aktuální, žádné nové migrace k aplikování.")
	} else {
		fmt.Printf("Úspěšně aplikováno %d nových migrací!\n", appliedCount)
	}
}