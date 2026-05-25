package utils

import "os"

func GetMigrationAuthor() string {
	author := os.Getenv("MIGRATION_AUTHOR")
	if author != "" {
		return author
	}

	// Linux/Mac
	author = os.Getenv("USER")
	if author == "" {
		//  Windows
		author = os.Getenv("USERNAME")
	}

	if author == "" {
		return "unknown_author"
	}
	return author
}