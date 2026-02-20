package journal

import (
	"database/sql"
	"fmt"
	"time"
)

// Store handles journal persistence in SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates a journal store using the provided database connection.
func NewStore(db *sql.DB) (*Store, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS journal_notes (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			date       TEXT NOT NULL UNIQUE,
			hidden     INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS journal_entries (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			note_id    INTEGER NOT NULL REFERENCES journal_notes(id) ON DELETE CASCADE,
			body       TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_journal_entries_note ON journal_entries(note_id)`,
		`CREATE INDEX IF NOT EXISTS idx_journal_notes_date ON journal_notes(date)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("journal migrate: %w", err)
		}
	}
	return nil
}

// ListNotes returns notes ordered by date descending.
// If includeHidden is false, hidden notes are excluded.
func (s *Store) ListNotes(includeHidden bool) ([]Note, error) {
	query := `SELECT id, date, hidden, created_at, updated_at FROM journal_notes`
	if !includeHidden {
		query += ` WHERE hidden = 0`
	}
	query += ` ORDER BY date DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		var dateStr, createdAt, updatedAt string
		var hidden int
		if err := rows.Scan(&n.ID, &dateStr, &hidden, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		n.Hidden = hidden != 0
		d, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
		if err != nil {
			return nil, fmt.Errorf("parse note date %q: %w", dateStr, err)
		}
		n.Date = d
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse note created_at %q: %w", createdAt, err)
		}
		n.CreatedAt = t
		t, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse note updated_at %q: %w", updatedAt, err)
		}
		n.UpdatedAt = t
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Batch-load all entries to avoid N+1 queries.
	if len(notes) > 0 {
		entryMap, err := s.listAllEntries(notes)
		if err != nil {
			return nil, fmt.Errorf("batch list entries: %w", err)
		}
		for i := range notes {
			notes[i].Entries = entryMap[notes[i].ID]
		}
	}
	return notes, nil
}

// GetOrCreateToday returns today's note, creating it if it does not exist.
func (s *Store) GetOrCreateToday() (*Note, error) {
	now := time.Now()
	dateStr := now.Format(time.DateOnly)
	nowStr := now.Format(time.RFC3339)

	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO journal_notes (date, hidden, created_at, updated_at) VALUES (?,0,?,?)`,
		dateStr, nowStr, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("create today note: %w", err)
	}

	var n Note
	var dStr, createdAt, updatedAt string
	var hidden int
	err = s.db.QueryRow(
		`SELECT id, date, hidden, created_at, updated_at FROM journal_notes WHERE date = ?`, dateStr,
	).Scan(&n.ID, &dStr, &hidden, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("get today note: %w", err)
	}
	n.Hidden = hidden != 0
	d, err := time.ParseInLocation(time.DateOnly, dStr, time.Local)
	if err != nil {
		return nil, fmt.Errorf("parse today note date %q: %w", dStr, err)
	}
	n.Date = d
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse today note created_at %q: %w", createdAt, err)
	}
	n.CreatedAt = t
	t, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse today note updated_at %q: %w", updatedAt, err)
	}
	n.UpdatedAt = t

	n.Entries, err = s.ListEntries(n.ID)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// AddEntry appends a new entry to the given note.
func (s *Store) AddEntry(noteID int64, body string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Format(time.RFC3339)
	if _, err := tx.Exec(
		`INSERT INTO journal_entries (note_id, body, created_at) VALUES (?,?,?)`,
		noteID, body, now,
	); err != nil {
		return fmt.Errorf("add entry: %w", err)
	}
	if _, err := tx.Exec(`UPDATE journal_notes SET updated_at = ? WHERE id = ?`, now, noteID); err != nil {
		return fmt.Errorf("update note timestamp: %w", err)
	}
	return tx.Commit()
}

// UpdateEntry replaces the body of an existing entry and updates the parent note's timestamp.
func (s *Store) UpdateEntry(entryID int64, body string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Format(time.RFC3339)
	res, err := tx.Exec(`UPDATE journal_entries SET body = ? WHERE id = ?`, body, entryID)
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("entry %d not found", entryID)
	}
	if _, err := tx.Exec(
		`UPDATE journal_notes SET updated_at = ? WHERE id = (SELECT note_id FROM journal_entries WHERE id = ?)`,
		now, entryID,
	); err != nil {
		return fmt.Errorf("update note timestamp: %w", err)
	}
	return tx.Commit()
}

// DeleteEntry removes a single entry and updates the parent note's timestamp.
func (s *Store) DeleteEntry(entryID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var noteID int64
	if err := tx.QueryRow(`SELECT note_id FROM journal_entries WHERE id = ?`, entryID).Scan(&noteID); err != nil {
		return fmt.Errorf("find entry note: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM journal_entries WHERE id = ?`, entryID); err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	now := time.Now().Format(time.RFC3339)
	if _, err := tx.Exec(`UPDATE journal_notes SET updated_at = ? WHERE id = ?`, now, noteID); err != nil {
		return fmt.Errorf("update note timestamp: %w", err)
	}
	return tx.Commit()
}

// ToggleHidden flips the hidden flag on a note.
func (s *Store) ToggleHidden(noteID int64) error {
	res, err := s.db.Exec(`UPDATE journal_notes SET hidden = NOT hidden WHERE id = ?`, noteID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("note %d not found", noteID)
	}
	return nil
}

// listAllEntries batch-loads entries for all given notes in a single query.
func (s *Store) listAllEntries(notes []Note) (map[int64][]Entry, error) {
	rows, err := s.db.Query(
		`SELECT id, note_id, body, created_at FROM journal_entries ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	noteIDs := make(map[int64]bool, len(notes))
	for _, n := range notes {
		noteIDs[n.ID] = true
	}

	result := make(map[int64][]Entry, len(notes))
	for rows.Next() {
		var e Entry
		var createdAt string
		if err := rows.Scan(&e.ID, &e.NoteID, &e.Body, &createdAt); err != nil {
			return nil, err
		}
		if !noteIDs[e.NoteID] {
			continue
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse entry created_at %q: %w", createdAt, err)
		}
		e.CreatedAt = t
		result[e.NoteID] = append(result[e.NoteID], e)
	}
	return result, rows.Err()
}

// RestoreEntry re-inserts a previously deleted journal entry.
func (s *Store) RestoreEntry(noteID int64, body string, createdAt time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO journal_entries (note_id, body, created_at) VALUES (?,?,?)`,
		noteID, body, createdAt.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("restore entry: %w", err)
	}
	now := time.Now().Format(time.RFC3339)
	if _, err := tx.Exec(`UPDATE journal_notes SET updated_at = ? WHERE id = ?`, now, noteID); err != nil {
		return fmt.Errorf("update note timestamp: %w", err)
	}
	return tx.Commit()
}

// ListEntries returns all entries for a note, ordered by created_at ASC.
func (s *Store) ListEntries(noteID int64) ([]Entry, error) {
	rows, err := s.db.Query(
		`SELECT id, note_id, body, created_at FROM journal_entries WHERE note_id = ? ORDER BY created_at ASC`,
		noteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		var createdAt string
		if err := rows.Scan(&e.ID, &e.NoteID, &e.Body, &createdAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse entry created_at %q: %w", createdAt, err)
		}
		e.CreatedAt = t
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
