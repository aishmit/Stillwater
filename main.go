package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbDir      = ".downloads_lifecycle"
	dbFile     = "classifications.db"
	downloads  = "Downloads"
	checkEvery = 10 * time.Second
)

func getDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, dbDir, dbFile)
}

func ensureDBSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE NOT NULL,
		classification INTEGER NOT NULL DEFAULT 0,
		classified_at TIMESTAMP,
		last_modified TIMESTAMP
	);`
	_, err := db.Exec(query)
	return err
}

func insertUnclassified(db *sql.DB, path string, modTime time.Time) error {
	query := `
	INSERT INTO files (path, classification, last_modified)
	VALUES (?, 0, ?)
	ON CONFLICT(path) DO UPDATE SET
		last_modified = excluded.last_modified
	`
	_, err := db.Exec(query, path, modTime, time.Now())
	return err
}

func scanDownloads(db *sql.DB) error {
	home, _ := os.UserHomeDir()
	downloadsPath := filepath.Join(home, downloads)
	fmt.Println("Scanning Downloads")

	return filepath.Walk(downloadsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		return insertUnclassified(db, path, info.ModTime())
	})
}

func main() {
	// Ensure DB path exists
	home, _ := os.UserHomeDir()
	dbPath := getDBPath()

	fmt.Println("Creating database directory if not exists.")

	os.MkdirAll(filepath.Join(home, dbDir), 0755)

	// Open DB
	fmt.Println("Creating database if not exists.")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Creating table if not exists.")

	if err := ensureDBSchema(db); err != nil {
		panic(err)
	}

	fmt.Println("Watching Downloads folder... (CTRL+C to stop)")

	for {
		err := scanDownloads(db)
		if err != nil {
			fmt.Println("Scan error:", err)
		}
		time.Sleep(checkEvery)
	}
}
