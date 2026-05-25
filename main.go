package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"cosmic/services"
)

var dbUrl string
var migDir string

var rootCmd = &cobra.Command{
	Use:   "cosmic",
	Short: "Cosmic je moderní a rychlý migrační nástroj",
	Long:  `Cosmic je Zero-Dependency databázový migrační nástroj napsaný v Go. Podporuje transakce, validaci integrity a zamykání.`,
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Aplikuje všechny čekající migrace do databáze",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Spouštím aplikaci migrací...")
		services.RunUp(dbUrl, migDir)
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Vrátí zpět poslední úspěšnou migraci (rollback)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Spouštím rollback...")
		services.RunDown(dbUrl, migDir)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Zobrazí aktuální stav všech migrací v databázi",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Zjišťuji stav databáze...")
		services.RunStatus(dbUrl, migDir)
	},
}

var createCmd = &cobra.Command{
	Use:   "create [název_migrace]",
	Short: "Vytvoří nové prázdné migrační soubory",
	//zkontroluje, že user napsal přesně 1 slovo
	Args: cobra.ExactArgs(1), 
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]
		fmt.Printf("📝 Vytvářím nové migrační soubory pro: %s...\n", migrationName)
		services.RunCreate(migDir, migrationName)
	},
}

func main() {
	godotenv.Load()

	defaultDB := os.Getenv("DATABASE_URL")
	defaultDir := os.Getenv("MIGRATIONS_DIR")
	
	if defaultDB == "" {
		defaultDB = "postgres://postgres:admin@localhost:5432/cosmic?sslmode=disable"
	}
	if defaultDir == "" {
		defaultDir = "./migrations"
	}	

	rootCmd.PersistentFlags().StringVar(&dbUrl, "db", defaultDB, "Připojovací řetězec k databázi")
	rootCmd.PersistentFlags().StringVar(&migDir, "dir", defaultDir, "Cesta ke složce s SQL migracemi")

	rootCmd.AddCommand(upCmd, downCmd, statusCmd, createCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
