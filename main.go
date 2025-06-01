package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	classificationQueue = make(chan string, 100)
	queuedFiles         = make(map[string]bool)
)

const (
	dbDir      = ".downloads_lifecycle"
	dbFile     = "classifications.db"
	downloads  = "Downloads"
	checkEvery = 10 * time.Second
)

type Classification int

const (
	NeverArchive    Classification = 1
	ArchiveAfter60d Classification = 2
	DeleteAfter60d  Classification = 3
)

func isValidClassification(input int) bool {
	switch Classification(input) {
	case NeverArchive, ArchiveAfter60d, DeleteAfter60d:
		return true
	default:
		return false
	}
}

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

func insertUnclassifiedInDB(db *sql.DB, path string, modTime time.Time) error {
	query := `
	INSERT INTO files (path, classification, last_modified)
	VALUES (?, 0, ?)
	ON CONFLICT(path) DO UPDATE SET
		last_modified = excluded.last_modified
	`
	_, err := db.Exec(query, path, modTime)
	return err
}

func updateClassificationInDB(db *sql.DB, path string, classification Classification) error {
	query := `
	INSERT INTO files (path, classification, classified_at)
	VALUES (?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		classification = excluded.classification,
		classified_at = excluded.classified_at
	`

	_, err := db.Exec(query, path, classification, time.Now())

	return err
}
func scanDownloads(db *sql.DB) ([]string, error) {
	home, _ := os.UserHomeDir()
	downloadsPath := filepath.Join(home, downloads)
	fmt.Println("Scanning Downloads")

	scannedFiles := []string{}

	err := filepath.Walk(downloadsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files and hidden directories
		name := info.Name()
		if strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir // Skip entire hidden directory subtree
			}
			return nil // Skip hidden file
		}

		if info.IsDir() {
			return nil // Continue walking
		}
		insertErr := insertUnclassifiedInDB(db, path, info.ModTime())
		if insertErr == nil {
			scannedFiles = append(scannedFiles, path)
		}
		return insertErr
	})

	return scannedFiles, err
}

func watcherLoop(db *sql.DB) {
	for {
		files, err := scanDownloads(db)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		for _, f := range files {
			if !queuedFiles[f] {
				queuedFiles[f] = true
				classificationQueue <- f
			}
		}
		time.Sleep(checkEvery)
	}
}

func classifierLoop(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	for file := range classificationQueue {
		for {
			fmt.Printf("\nClassify '%s':\n[1] Never archive\n[2] Archive after 60 days\n[3] Delete after 60 days\n> ", file)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			castInput, err := strconv.Atoi(input)

			if err == nil && isValidClassification(castInput) {
				classification := Classification(castInput)
				updateClassificationInDB(db, file, classification)
				fmt.Printf("✅ File '%s' classified as %d\n", file, int(classification))
				break
			} else {
				fmt.Println("⚠️ Invalid input. Please enter 1, 2, or 3.")
			}
		}
	}
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

	go watcherLoop(db)
	classifierLoop(db)
}
