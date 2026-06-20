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
	Short: "Cosmic is a modern and fast migration tool",
	Long:  `Cosmic is a zero-dependency database migration tool written in Go. It supports transactions, integrity validation, and locking.`,
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Applies all pending migrations to the database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting migration application...")
		services.RunUp(dbUrl, migDir)
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rolls back the last successful migration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting rollback...")
		services.RunDown(dbUrl, migDir)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Displays the current status of all database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking database status...")
		services.RunStatus(dbUrl, migDir)
	},
}

var createCmd = &cobra.Command{
	Use:   "create [migration_name]",
	Short: "Creates new empty migration files",

	// zkontroluje, že user napsal přesně 1 slovo
	Args: cobra.ExactArgs(1), 
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]
		fmt.Printf("Creating new migration files for: %s...\n", migrationName)
		services.RunCreate(migDir, migrationName)
	},
}

func main() {
	godotenv.Load()

	defaultDB := os.Getenv("DATABASE_URL")
	defaultDir := os.Getenv("MIGRATIONS_DIR")
	
	if defaultDB == "" {
		// defaultDB = "postgres://postgres:admin@localhost:5432/cosmic?sslmode=disable"
		fmt.Println("Warning: DATABASE_URL is empty")
	}
	if defaultDir == "" {
		defaultDir = "./migrations"
	}   

	rootCmd.PersistentFlags().StringVar(&dbUrl, "db", defaultDB, "Database connection string")
	rootCmd.PersistentFlags().StringVar(&migDir, "dir", defaultDir, "Path to the SQL migrations directory")

	rootCmd.AddCommand(upCmd, downCmd, statusCmd, createCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}