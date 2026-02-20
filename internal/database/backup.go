package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Backup creates a backup of the SQLite database using VACUUM INTO.
// The backup file is named backup-YYYY-MM-DD.db and placed in dir.
// Any backup files older than retainDays are pruned. The directory
// is created if it does not already exist.
func Backup(db *sql.DB, dir string, retainDays int) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	name := fmt.Sprintf("backup-%s.db", time.Now().Format("2006-01-02"))
	dest := filepath.Join(dir, name)

	// Skip if today's backup already exists.
	if _, err := os.Stat(dest); err == nil {
		return pruneBackups(dir, retainDays)
	}

	// VACUUM INTO creates a standalone copy of the database.
	if _, err := db.Exec(fmt.Sprintf(`VACUUM INTO '%s'`, dest)); err != nil {
		return fmt.Errorf("vacuum into %s: %w", dest, err)
	}

	// Prune old backups.
	if err := pruneBackups(dir, retainDays); err != nil {
		return fmt.Errorf("prune backups: %w", err)
	}

	return nil
}

// pruneBackups removes backup-YYYY-MM-DD.db files from dir that are
// older than retainDays days.
func pruneBackups(dir string, retainDays int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -retainDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "backup-") || !strings.HasSuffix(name, ".db") {
			continue
		}

		// Extract date from backup-YYYY-MM-DD.db
		dateStr := strings.TrimPrefix(name, "backup-")
		dateStr = strings.TrimSuffix(dateStr, ".db")

		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue // not a valid backup file name, skip
		}

		if t.Before(cutoff) {
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("remove %s: %w", path, err)
			}
		}
	}

	return nil
}
