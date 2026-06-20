package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func RunCreate(migrationsDir string, name string) {

	// formát YYYYMMDDHHMMSS
	timestamp := time.Now().Format("20060102150405")

	baseName := fmt.Sprintf("%s_%s", timestamp, name)
	upFile := filepath.Join(migrationsDir, baseName+".up.sql")
	downFile := filepath.Join(migrationsDir, baseName+".down.sql")

	// Pokud složka migrations náhodou neexistuje, nástroj ji sám vytvoří
	err := os.MkdirAll(migrationsDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Vytvoření prázdných souborů na disku
	// 0644 jsou standardní práva (čtení a zápis pro tebe, čtení pro ostatní)
	err = os.WriteFile(upFile, []byte(""), 0644)
	if err != nil {
		log.Fatalf("Failed to create UP file: %v", err)
	}

	err = os.WriteFile(downFile, []byte(""), 0644)
	if err != nil {
		log.Fatalf("Failed to create DOWN file: %v", err)
	}

	fmt.Printf("Successfully created migration files:\n")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)
}