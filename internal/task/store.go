package task

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db *sql.DB
}

// NewStore creates a task store using the provided database connection.
// The caller is responsible for opening and closing the DB.
func NewStore(db *sql.DB) (*Store, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status      INTEGER NOT NULL DEFAULT 0,
			priority    INTEGER NOT NULL DEFAULT 0,
			due_date    TEXT,
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS subtasks (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			title      TEXT NOT NULL,
			completed  INTEGER NOT NULL DEFAULT 0,
			position   INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id      INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			name    TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_subtasks_task ON subtasks(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tags_task ON tags(task_id)`,
		`CREATE TABLE IF NOT EXISTS time_logs (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id   INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			duration  INTEGER NOT NULL,
			note      TEXT NOT NULL DEFAULT '',
			logged_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_time_logs_task ON time_logs(task_id)`,
		`CREATE TABLE IF NOT EXISTS task_dependencies (
			task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			blocked_by INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			PRIMARY KEY (task_id, blocked_by)
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	// Add new columns to tasks table if they don't already exist.
	if err := addColumnIfNotExists(db, "tasks", "recur_freq", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("migrate recur_freq: %w", err)
	}
	if err := addColumnIfNotExists(db, "tasks", "recur_interval", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("migrate recur_interval: %w", err)
	}
	return nil
}

func addColumnIfNotExists(db *sql.DB, table, column, colDef string) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil // column already exists
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, colDef))
	return err
}

func (s *Store) List() ([]Task, error) {
	rows, err := s.db.Query(`SELECT id, title, description, status, priority, due_date, created_at, updated_at, recur_freq, recur_interval FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var dueDate, createdAt, updatedAt sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &dueDate, &createdAt, &updatedAt, &t.RecurFreq, &t.RecurInterval); err != nil {
			return nil, err
		}
		if dueDate.Valid {
			d, err := time.ParseInLocation(time.DateOnly, dueDate.String, time.Local)
			if err != nil {
				return nil, fmt.Errorf("parse task due_date %q: %w", dueDate.String, err)
			}
			t.DueDate = &d
		}
		if createdAt.Valid {
			parsed, err := time.Parse(time.RFC3339, createdAt.String)
			if err != nil {
				return nil, fmt.Errorf("parse task created_at %q: %w", createdAt.String, err)
			}
			t.CreatedAt = parsed
		}
		if updatedAt.Valid {
			parsed, err := time.Parse(time.RFC3339, updatedAt.String)
			if err != nil {
				return nil, fmt.Errorf("parse task updated_at %q: %w", updatedAt.String, err)
			}
			t.UpdatedAt = parsed
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range tasks {
		if tasks[i].Subtasks, err = s.listSubtasks(tasks[i].ID); err != nil {
			return nil, fmt.Errorf("list subtasks for task %d: %w", tasks[i].ID, err)
		}
		if tasks[i].Tags, err = s.listTags(tasks[i].ID); err != nil {
			return nil, fmt.Errorf("list tags for task %d: %w", tasks[i].ID, err)
		}
		if tasks[i].TimeLogs, err = s.ListTimeLogs(tasks[i].ID); err != nil {
			return nil, fmt.Errorf("list time logs for task %d: %w", tasks[i].ID, err)
		}
		if tasks[i].BlockedByIDs, err = s.ListBlockerIDs(tasks[i].ID); err != nil {
			return nil, fmt.Errorf("list blocker ids for task %d: %w", tasks[i].ID, err)
		}
	}
	return tasks, nil
}

func (s *Store) listSubtasks(taskID int64) ([]Subtask, error) {
	rows, err := s.db.Query(`SELECT id, title, completed, position FROM subtasks WHERE task_id = ? ORDER BY position`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []Subtask
	for rows.Next() {
		var st Subtask
		if err := rows.Scan(&st.ID, &st.Title, &st.Completed, &st.Position); err != nil {
			return nil, err
		}
		subs = append(subs, st)
	}
	return subs, rows.Err()
}

func (s *Store) listTags(taskID int64) ([]string, error) {
	rows, err := s.db.Query(`SELECT name FROM tags WHERE task_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
}

func (s *Store) Create(t *Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	var dueStr *string
	if t.DueDate != nil {
		d := t.DueDate.Format(time.DateOnly)
		dueStr = &d
	}
	res, err := tx.Exec(
		`INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		t.Title, t.Description, t.Status, t.Priority, dueStr, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	t.ID = id

	if err := saveTagsTx(tx, t.ID, t.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

func saveTagsTx(tx *sql.Tx, taskID int64, tags []string) error {
	if _, err := tx.Exec(`DELETE FROM tags WHERE task_id = ?`, taskID); err != nil {
		return fmt.Errorf("delete tags: %w", err)
	}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO tags (task_id, name) VALUES (?,?)`, taskID, tag); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Update(t *Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	t.UpdatedAt = time.Now()
	var dueStr *string
	if t.DueDate != nil {
		d := t.DueDate.Format(time.DateOnly)
		dueStr = &d
	}
	if _, err := tx.Exec(
		`UPDATE tasks SET title=?, description=?, status=?, priority=?, due_date=?, updated_at=? WHERE id=?`,
		t.Title, t.Description, t.Status, t.Priority, dueStr, t.UpdatedAt.Format(time.RFC3339), t.ID,
	); err != nil {
		return err
	}

	if err := saveTagsTx(tx, t.ID, t.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func (s *Store) AddSubtask(taskID int64, title string) error {
	var maxPos int
	if err := s.db.QueryRow(`SELECT COALESCE(MAX(position), -1) FROM subtasks WHERE task_id = ?`, taskID).Scan(&maxPos); err != nil {
		return fmt.Errorf("get max position: %w", err)
	}
	_, err := s.db.Exec(`INSERT INTO subtasks (task_id, title, position) VALUES (?,?,?)`, taskID, title, maxPos+1)
	return err
}

func (s *Store) ToggleSubtask(id int64) error {
	_, err := s.db.Exec(`UPDATE subtasks SET completed = NOT completed WHERE id = ?`, id)
	return err
}

func (s *Store) UpdateSubtask(id int64, title string) error {
	_, err := s.db.Exec(`UPDATE subtasks SET title = ? WHERE id = ?`, title, id)
	return err
}

func (s *Store) DeleteSubtask(id int64) error {
	_, err := s.db.Exec(`DELETE FROM subtasks WHERE id = ?`, id)
	return err
}

// AddTimeLog records a time log entry for the given task.
func (s *Store) AddTimeLog(taskID int64, duration time.Duration, note string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO time_logs (task_id, duration, note, logged_at) VALUES (?,?,?,?)`,
		taskID, int64(duration), note, now,
	)
	return err
}

// ListTimeLogs returns all time logs for a task, ordered by logged_at descending.
func (s *Store) ListTimeLogs(taskID int64) ([]TimeLog, error) {
	rows, err := s.db.Query(
		`SELECT id, task_id, duration, note, logged_at FROM time_logs WHERE task_id = ? ORDER BY logged_at DESC`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []TimeLog
	for rows.Next() {
		var tl TimeLog
		var dur int64
		var loggedAt string
		if err := rows.Scan(&tl.ID, &tl.TaskID, &dur, &tl.Note, &loggedAt); err != nil {
			return nil, err
		}
		tl.Duration = time.Duration(dur)
		parsed, err := time.Parse(time.RFC3339, loggedAt)
		if err != nil {
			return nil, fmt.Errorf("parse time_log logged_at %q: %w", loggedAt, err)
		}
		tl.LoggedAt = parsed
		logs = append(logs, tl)
	}
	return logs, rows.Err()
}

// SetBlocker adds a dependency: taskID is blocked by blockerID.
func (s *Store) SetBlocker(taskID, blockerID int64) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO task_dependencies (task_id, blocked_by) VALUES (?,?)`,
		taskID, blockerID,
	)
	return err
}

// RemoveBlocker removes a dependency: taskID is no longer blocked by blockerID.
func (s *Store) RemoveBlocker(taskID, blockerID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM task_dependencies WHERE task_id = ? AND blocked_by = ?`,
		taskID, blockerID,
	)
	return err
}

// ListBlockerIDs returns all task IDs that block the given task.
func (s *Store) ListBlockerIDs(taskID int64) ([]int64, error) {
	rows, err := s.db.Query(
		`SELECT blocked_by FROM task_dependencies WHERE task_id = ?`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// UpdateRecurrence sets the recurrence frequency and interval for a task.
func (s *Store) UpdateRecurrence(taskID int64, freq RecurFreq, interval int) error {
	_, err := s.db.Exec(
		`UPDATE tasks SET recur_freq = ?, recur_interval = ? WHERE id = ?`,
		int(freq), interval, taskID,
	)
	return err
}

// Restore re-inserts a previously deleted task with its tags.
func (s *Store) Restore(t *Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var dueStr *string
	if t.DueDate != nil {
		d := t.DueDate.Format(time.DateOnly)
		dueStr = &d
	}
	res, err := tx.Exec(
		`INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at, recur_freq, recur_interval) VALUES (?,?,?,?,?,?,?,?,?)`,
		t.Title, t.Description, t.Status, t.Priority, dueStr,
		t.CreatedAt.Format(time.RFC3339), t.UpdatedAt.Format(time.RFC3339),
		int(t.RecurFreq), t.RecurInterval,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id

	if err := saveTagsTx(tx, t.ID, t.Tags); err != nil {
		return err
	}
	for _, st := range t.Subtasks {
		if _, err := tx.Exec(
			`INSERT INTO subtasks (task_id, title, completed, position) VALUES (?,?,?,?)`,
			t.ID, st.Title, st.Completed, st.Position,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RestoreSubtask re-inserts a previously deleted subtask.
func (s *Store) RestoreSubtask(taskID int64, title string, completed bool, position int) error {
	_, err := s.db.Exec(
		`INSERT INTO subtasks (task_id, title, completed, position) VALUES (?,?,?,?)`,
		taskID, title, completed, position,
	)
	return err
}

// GetByID retrieves a single task by ID.
func (s *Store) GetByID(id int64) (*Task, error) {
	var t Task
	var dueDate, createdAt, updatedAt sql.NullString
	err := s.db.QueryRow(
		`SELECT id, title, description, status, priority, due_date, created_at, updated_at, recur_freq, recur_interval FROM tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &dueDate, &createdAt, &updatedAt, &t.RecurFreq, &t.RecurInterval)
	if err != nil {
		return nil, err
	}
	if dueDate.Valid {
		d, _ := time.ParseInLocation(time.DateOnly, dueDate.String, time.Local)
		t.DueDate = &d
	}
	if createdAt.Valid {
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}
	if updatedAt.Valid {
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
	}
	t.Subtasks, _ = s.listSubtasks(t.ID)
	t.Tags, _ = s.listTags(t.ID)
	t.TimeLogs, _ = s.ListTimeLogs(t.ID)
	t.BlockedByIDs, _ = s.ListBlockerIDs(t.ID)
	return &t, nil
}
