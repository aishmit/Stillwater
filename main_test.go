package main

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if err := ensureDBSchema(db); err != nil {
		t.Fatalf("Failed to setup schema: %v", err)
	}
	return db
}

func TestIsValidClassification(t *testing.T) {
	tests := []struct {
		input    int
		expected bool
	}{
		{1, true},
		{2, true},
		{3, true},
		{4, false},
		{0, false},
		{-1, false},
	}

	for _, tt := range tests {
		result := isValidClassification(tt.input)
		if result != tt.expected {
			t.Errorf("Expected isValidClassification(%d) = %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestInsertAndUpdateClassificationInDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	filePath := "/tmp/testfile.txt"
	modTime := time.Now()

	err := insertUnclassifiedInDB(db, filePath, modTime)
	if err != nil {
		t.Fatalf("insertUnclassifiedInDB failed: %v", err)
	}

	// Check that the file was inserted with classification = 0
	var classification int
	err = db.QueryRow("SELECT classification FROM files WHERE path = ?", filePath).Scan(&classification)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if classification != 0 {
		t.Errorf("Expected classification 0, got %d", classification)
	}

	// Update classification
	err = updateClassificationInDB(db, filePath, ArchiveAfter60d)
	if err != nil {
		t.Fatalf("updateClassificationInDB failed: %v", err)
	}

	err = db.QueryRow("SELECT classification FROM files WHERE path = ?", filePath).Scan(&classification)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if classification != int(ArchiveAfter60d) {
		t.Errorf("Expected classification %d, got %d", int(ArchiveAfter60d), classification)
	}
}

func TestEnsureDBSchemaCreatesTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Query("SELECT * FROM files LIMIT 1")
	if err != nil {
		t.Errorf("Table 'files' should exist but query failed: %v", err)
	}
}
