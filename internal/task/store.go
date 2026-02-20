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
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
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

	for i := range tasks {
		if tasks[i].Subtasks, err = s.listSubtasks(tasks[i].ID); err != nil {
			return nil, fmt.Errorf("list subtasks for task %d: %w", tasks[i].ID, err)
		}
		if tasks[i].Tags, err = s.listTags(tasks[i].ID); err != nil {
			return nil, fmt.Errorf("list tags for task %d: %w", tasks[i].ID, err)
		}
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
