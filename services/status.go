package services

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

func RunStatus(dbUrl string, migrationsDir string) {
	db := ConnectDB(dbUrl)
	defer db.Close()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cosmic_schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		hash VARCHAR(255),
		author VARCHAR(255)
	)`)
	if err != nil {
		log.Fatalf("Nelze ověřit stavovou tabulku: %v", err)
	}

	// Místo struktury vytvoříme dvě jednoduché mapy
	appliedTimes := make(map[string]string)
	appliedAuthors := make(map[string]string)

	rows, err := db.Query("SELECT version, applied_at, author FROM cosmic_schema_migrations")
	if err != nil {
		log.Fatalf("Chyba při čtení historie z databáze: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		var appliedAt string
		var author string
		if err := rows.Scan(&version, &appliedAt, &author); err != nil {
			log.Fatalf("Chyba při parsování řádku z databáze: %v", err)
		}
		// Uloží data do příslušných map
		appliedTimes[version] = appliedAt
		appliedAuthors[version] = author
	}

	files, err := os.ReadDir(migrationsDir)
	var allVersions []string
	diskVersions := make(map[string]bool)

	if err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
				version := strings.TrimSuffix(f.Name(), ".up.sql")
				allVersions = append(allVersions, version)
				diskVersions[version] = true
			}
		}
	} else {
		fmt.Printf("Upozornění: Složka %s neexistuje nebo ji nelze přečíst.\n", migrationsDir)
	}

	for dbVersion := range appliedTimes {
		if !diskVersions[dbVersion] {
			allVersions = append(allVersions, dbVersion)
		}
	}

	allVersions = uniqueAndSort(allVersions)

	fmt.Println("\n STAV MIGRACÍ DATABÁZE")
	fmt.Println("==========================================================================================")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "VERZE\tSTAV\tAUTOR\tAPLIKOVÁNO KDY\t")
	fmt.Fprintln(w, "-----\t----\t-----\t--------------\t")

	for _, version := range allVersions {
		status := ""
		appliedAt := "-"
		author := "-"

		// Zkontrolujeme, jestli verze existuje v mapě aplikovaných časů
		timeVal, isApplied := appliedTimes[version]

		if isApplied {
			if diskVersions[version] {
				status = "Aplikováno"
			} else {
				status = "Chybí soubor"
			}
			appliedAt = strings.Split(timeVal, ".")[0]
			author = appliedAuthors[version] // Získá autora z druhé mapy
		} else {
			status = "Čeká"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", version, status, author, appliedAt)
	}
	w.Flush()
	fmt.Println("==========================================================================================\n")
}

func uniqueAndSort(strSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	sort.Strings(list)
	return list
}