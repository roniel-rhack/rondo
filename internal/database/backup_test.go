package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// newTestDB creates a temporary in-memory-like SQLite database with a table.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	// Use a temp file because VACUUM INTO does not work with :memory:.
	tmp := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", tmp)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO t (val) VALUES ('hello')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestBackup_CreatesFile(t *testing.T) {
	db := newTestDB(t)
	dir := filepath.Join(t.TempDir(), "backups")

	if err := Backup(db, dir, 7); err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	expected := fmt.Sprintf("backup-%s.db", time.Now().Format("2006-01-02"))
	path := filepath.Join(dir, expected)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("backup file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("backup file is empty")
	}

	// Verify the backup is a valid SQLite database with our data.
	bdb, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	defer bdb.Close()

	var val string
	if err := bdb.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val); err != nil {
		t.Fatalf("query backup: %v", err)
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}
}

func TestBackup_CreatesDirectory(t *testing.T) {
	db := newTestDB(t)
	dir := filepath.Join(t.TempDir(), "nested", "deep", "backups")

	if err := Backup(db, dir, 7); err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("directory not created: %v", err)
	}
}

func TestBackup_PrunesOldFiles(t *testing.T) {
	db := newTestDB(t)
	dir := t.TempDir()

	// Create fake old backup files.
	oldDate := time.Now().AddDate(0, 0, -10)
	recentDate := time.Now().AddDate(0, 0, -2)

	oldFile := filepath.Join(dir, fmt.Sprintf("backup-%s.db", oldDate.Format("2006-01-02")))
	recentFile := filepath.Join(dir, fmt.Sprintf("backup-%s.db", recentDate.Format("2006-01-02")))
	unrelatedFile := filepath.Join(dir, "not-a-backup.txt")

	for _, f := range []string{oldFile, recentFile, unrelatedFile} {
		if err := os.WriteFile(f, []byte("data"), 0o644); err != nil {
			t.Fatalf("create file: %v", err)
		}
	}

	if err := Backup(db, dir, 7); err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	// Old file should be pruned.
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("old backup should have been pruned: %s", oldFile)
	}

	// Recent file should remain.
	if _, err := os.Stat(recentFile); err != nil {
		t.Errorf("recent backup should still exist: %v", err)
	}

	// Unrelated file should remain.
	if _, err := os.Stat(unrelatedFile); err != nil {
		t.Errorf("unrelated file should still exist: %v", err)
	}

	// Today's backup should exist.
	todayFile := filepath.Join(dir, fmt.Sprintf("backup-%s.db", time.Now().Format("2006-01-02")))
	if _, err := os.Stat(todayFile); err != nil {
		t.Errorf("today's backup should exist: %v", err)
	}
}
