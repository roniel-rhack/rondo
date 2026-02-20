package focus

import (
	"database/sql"
	"fmt"
	"time"
)

// Store handles focus session persistence in SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates a focus store using the provided database connection.
// The caller is responsible for opening and closing the DB.
func NewStore(db *sql.DB) (*Store, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS focus_sessions (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id      INTEGER NOT NULL DEFAULT 0,
			duration     INTEGER NOT NULL,
			started_at   TEXT NOT NULL,
			completed_at TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_focus_sessions_task ON focus_sessions(task_id)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("focus migrate: %w", err)
		}
	}
	return nil
}

// Create inserts a new session and sets its ID.
func (s *Store) Create(session *Session) error {
	var completedAt *string
	if session.CompletedAt != nil {
		v := session.CompletedAt.Format(time.RFC3339)
		completedAt = &v
	}
	res, err := s.db.Exec(
		`INSERT INTO focus_sessions (task_id, duration, started_at, completed_at) VALUES (?,?,?,?)`,
		session.TaskID,
		int64(session.Duration),
		session.StartedAt.Format(time.RFC3339),
		completedAt,
	)
	if err != nil {
		return fmt.Errorf("create focus session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	session.ID = id
	return nil
}

// Complete marks a session as completed by setting completed_at to now.
func (s *Store) Complete(id int64) error {
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(`UPDATE focus_sessions SET completed_at = ? WHERE id = ?`, now, id)
	if err != nil {
		return fmt.Errorf("complete focus session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("focus session %d not found", id)
	}
	return nil
}

// ListByTask returns sessions for a given task, ordered by started_at DESC.
func (s *Store) ListByTask(taskID int64) ([]Session, error) {
	rows, err := s.db.Query(
		`SELECT id, task_id, duration, started_at, completed_at FROM focus_sessions WHERE task_id = ? ORDER BY started_at DESC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("list focus sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

// CompletionsByDay returns the count of completed sessions per day for the
// last N days, keyed by "YYYY-MM-DD".
func (s *Store) CompletionsByDay(days int) (map[string]int, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)
	rows, err := s.db.Query(
		`SELECT DATE(completed_at) AS day, COUNT(*) FROM focus_sessions
		 WHERE completed_at IS NOT NULL AND completed_at >= ?
		 GROUP BY day`,
		cutoff,
	)
	if err != nil {
		return nil, fmt.Errorf("completions by day: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		result[day] = count
	}
	return result, rows.Err()
}

// TodayCount returns the number of sessions completed today.
func (s *Store) TodayCount() (int, error) {
	today := time.Now().Format(time.DateOnly)
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM focus_sessions WHERE completed_at IS NOT NULL AND DATE(completed_at) = ?`,
		today,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("today count: %w", err)
	}
	return count, nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanSession(s scanner) (Session, error) {
	var sess Session
	var durationNs int64
	var startedAt string
	var completedAt sql.NullString

	if err := s.Scan(&sess.ID, &sess.TaskID, &durationNs, &startedAt, &completedAt); err != nil {
		return Session{}, err
	}

	sess.Duration = time.Duration(durationNs)

	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return Session{}, fmt.Errorf("parse started_at %q: %w", startedAt, err)
	}
	sess.StartedAt = t

	if completedAt.Valid {
		t, err := time.Parse(time.RFC3339, completedAt.String)
		if err != nil {
			return Session{}, fmt.Errorf("parse completed_at %q: %w", completedAt.String, err)
		}
		sess.CompletedAt = &t
	}
	return sess, nil
}
