package task

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".todo-app")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}
	dbPath := filepath.Join(dir, "todo.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		title       TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status      INTEGER NOT NULL DEFAULT 0,
		priority    INTEGER NOT NULL DEFAULT 0,
		due_date    TEXT,
		created_at  TEXT NOT NULL,
		updated_at  TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS subtasks (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		title      TEXT NOT NULL,
		completed  INTEGER NOT NULL DEFAULT 0,
		position   INTEGER NOT NULL DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS tags (
		id      INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		name    TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_subtasks_task ON subtasks(task_id);
	CREATE INDEX IF NOT EXISTS idx_tags_task ON tags(task_id);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) List() ([]Task, error) {
	rows, err := s.db.Query(`SELECT id, title, description, status, priority, due_date, created_at, updated_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var dueDate, createdAt, updatedAt sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &dueDate, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if dueDate.Valid {
			if d, err := time.Parse(time.DateOnly, dueDate.String); err == nil {
				t.DueDate = &d
			}
		}
		if createdAt.Valid {
			t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		if updatedAt.Valid {
			t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
		}
		tasks = append(tasks, t)
	}

	for i := range tasks {
		tasks[i].Subtasks, _ = s.listSubtasks(tasks[i].ID)
		tasks[i].Tags, _ = s.listTags(tasks[i].ID)
	}
	return tasks, rows.Err()
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
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	var dueStr *string
	if t.DueDate != nil {
		d := t.DueDate.Format(time.DateOnly)
		dueStr = &d
	}
	res, err := s.db.Exec(
		`INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		t.Title, t.Description, t.Status, t.Priority, dueStr, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	t.ID, _ = res.LastInsertId()
	return s.saveTags(t.ID, t.Tags)
}

func (s *Store) saveTags(taskID int64, tags []string) error {
	s.db.Exec(`DELETE FROM tags WHERE task_id = ?`, taskID)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, err := s.db.Exec(`INSERT INTO tags (task_id, name) VALUES (?,?)`, taskID, tag); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Update(t *Task) error {
	t.UpdatedAt = time.Now()
	var dueStr *string
	if t.DueDate != nil {
		d := t.DueDate.Format(time.DateOnly)
		dueStr = &d
	}
	_, err := s.db.Exec(
		`UPDATE tasks SET title=?, description=?, status=?, priority=?, due_date=?, updated_at=? WHERE id=?`,
		t.Title, t.Description, t.Status, t.Priority, dueStr, t.UpdatedAt.Format(time.RFC3339), t.ID,
	)
	if err != nil {
		return err
	}
	return s.saveTags(t.ID, t.Tags)
}

func (s *Store) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func (s *Store) AddSubtask(taskID int64, title string) error {
	var maxPos int
	s.db.QueryRow(`SELECT COALESCE(MAX(position), -1) FROM subtasks WHERE task_id = ?`, taskID).Scan(&maxPos)
	_, err := s.db.Exec(`INSERT INTO subtasks (task_id, title, position) VALUES (?,?,?)`, taskID, title, maxPos+1)
	return err
}

func (s *Store) ToggleSubtask(id int64) error {
	_, err := s.db.Exec(`UPDATE subtasks SET completed = NOT completed WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteSubtask(id int64) error {
	_, err := s.db.Exec(`DELETE FROM subtasks WHERE id = ?`, id)
	return err
}
