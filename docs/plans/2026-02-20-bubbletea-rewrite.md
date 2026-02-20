# Todo App - Go + Bubbletea Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite the todo TUI app from Java/TamboUI to Go using the Charm ecosystem (Bubbletea + Bubbles + Lip Gloss + Huh) with SQLite persistence, replacing the existing Java project in-place.

**Architecture:** Bubbletea MVU (Model-Update-View) pattern. The main `model` struct owns a `list.Model` for the task list, a `viewport.Model` for the detail panel, and a `*huh.Form` for task add/edit dialogs. SQLite via `modernc.org/sqlite` (CGO-free) stores tasks, subtasks, and tags. Custom `list.ItemDelegate` renders tasks with priority colors and status icons.

**Tech Stack:** Go 1.23+, Bubbletea v1.3.x, Bubbles v1.0.x, Lip Gloss v1.1.x, Huh v0.8.x, modernc.org/sqlite

---

### Task 1: Scaffold Go project and clean up Java files

**Files:**
- Create: `go.mod`
- Create: `cmd/todo/main.go`
- Delete: `src/` (entire directory)
- Delete: `pom.xml`
- Delete: `target/` (if exists)

**Step 1: Remove Java project files**

```bash
rm -rf src/ pom.xml target/ .mvn mvnw mvnw.cmd
```

Keep `docs/`, `CLAUDE.md`, `.git/`, `.gitignore`, `data/`.

**Step 2: Initialize Go module**

```bash
go mod init github.com/roniel/todo-app
```

**Step 3: Create directory structure**

```bash
mkdir -p cmd/todo internal/app internal/task internal/ui
```

**Step 4: Create minimal main.go**

```go
// cmd/todo/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("todo app starting...")
	os.Exit(0)
}
```

**Step 5: Verify it compiles**

Run: `go build ./cmd/todo`
Expected: No errors, binary created.

**Step 6: Fetch dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/huh@latest
go get modernc.org/sqlite@latest
```

**Step 7: Verify go.sum is populated**

Run: `go mod tidy`
Expected: `go.sum` file generated with all transitive dependencies.

**Step 8: Commit**

```bash
git add -A
git commit -m "chore: scaffold Go project, remove Java files"
```

---

### Task 2: Domain model - Task, Subtask, Status, Priority

**Files:**
- Create: `internal/task/task.go`

**Step 1: Create the domain types**

```go
// internal/task/task.go
package task

import (
	"time"
)

type Status int

const (
	Pending    Status = iota
	InProgress
	Done
)

func (s Status) String() string {
	switch s {
	case InProgress:
		return "In Progress"
	case Done:
		return "Done"
	default:
		return "Pending"
	}
}

func (s Status) Icon() string {
	switch s {
	case InProgress:
		return "◐"
	case Done:
		return "✓"
	default:
		return "○"
	}
}

func (s Status) Next() Status {
	switch s {
	case Pending:
		return InProgress
	case InProgress:
		return Done
	default:
		return Pending
	}
}

type Priority int

const (
	Low Priority = iota
	Medium
	High
	Urgent
)

func (p Priority) String() string {
	switch p {
	case Medium:
		return "Medium"
	case High:
		return "High"
	case Urgent:
		return "Urgent"
	default:
		return "Low"
	}
}

func (p Priority) Label() string {
	switch p {
	case Medium:
		return "▪▪"
	case High:
		return "▪▪▪"
	case Urgent:
		return "!!!"
	default:
		return "▪"
	}
}

type Subtask struct {
	ID        int64
	Title     string
	Completed bool
	Position  int
}

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      Status
	Priority    Priority
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Subtasks    []Subtask
	Tags        []string
}

// FilterValue implements list.Item interface for bubbles list.
func (t Task) FilterValue() string { return t.Title }
```

**Step 2: Compile**

Run: `go build ./internal/task/...`
Expected: No errors.

**Step 3: Commit**

```bash
git add internal/task/task.go
git commit -m "feat: add domain model - Task, Subtask, Status, Priority"
```

---

### Task 3: SQLite store - schema and CRUD

**Files:**
- Create: `internal/task/store.go`

**Step 1: Create the Store with schema initialization**

```go
// internal/task/store.go
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
```

**Step 2: Add List method**

```go
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
```

**Step 3: Add Create method**

```go
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
```

**Step 4: Add Update method**

```go
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
```

**Step 5: Add Delete method**

```go
func (s *Store) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}
```

**Step 6: Add subtask methods**

```go
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
```

**Step 7: Compile**

Run: `go build ./internal/task/...`
Expected: No errors.

**Step 8: Commit**

```bash
git add internal/task/store.go
git commit -m "feat: add SQLite store with schema, CRUD, subtasks, tags"
```

---

### Task 4: Lip Gloss styles - theme definition

**Files:**
- Create: `internal/app/styles.go`

**Step 1: Create the styles file**

```go
// internal/app/styles.go
package app

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	cyan    = lipgloss.Color("#00BCD4")
	white   = lipgloss.Color("#FAFAFA")
	gray    = lipgloss.Color("#666666")
	dimGray = lipgloss.Color("#444444")
	black   = lipgloss.Color("#000000")
	red     = lipgloss.Color("#F44336")
	yellow  = lipgloss.Color("#FFC107")
	green   = lipgloss.Color("#4CAF50")
	magenta = lipgloss.Color("#E040FB")

	// App frame
	appStyle = lipgloss.NewStyle()

	// Header / tabs
	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(gray)

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(cyan).
			Bold(true).
			Reverse(true)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(dimGray)

	// Task list panel
	listPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			Padding(0, 1)

	listPanelFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cyan).
				Padding(0, 1)

	// Detail panel
	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(gray).
				Padding(1, 2)

	// Detail content
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(white)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(gray).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(white)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(gray)

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(cyan)

	// Priority colors
	priorityLowStyle    = lipgloss.NewStyle().Foreground(green)
	priorityMedStyle    = lipgloss.NewStyle().Foreground(yellow)
	priorityHighStyle   = lipgloss.NewStyle().Foreground(red)
	priorityUrgentStyle = lipgloss.NewStyle().Foreground(magenta)

	// Status icons
	statusPendingStyle    = lipgloss.NewStyle().Foreground(gray)
	statusInProgressStyle = lipgloss.NewStyle().Foreground(yellow)
	statusDoneStyle       = lipgloss.NewStyle().Foreground(green)

	// Dialog overlay
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(1, 2).
			Width(60)

	// Subtask styles
	subtaskDoneStyle   = lipgloss.NewStyle().Foreground(green).Strikethrough(true)
	subtaskUndoneStyle = lipgloss.NewStyle().Foreground(white)

	// Help
	helpKeyStyle  = lipgloss.NewStyle().Foreground(cyan)
	helpDescStyle = lipgloss.NewStyle().Foreground(gray)
)
```

**Step 2: Compile**

Run: `go build ./internal/app/...`
Expected: No errors.

**Step 3: Commit**

```bash
git add internal/app/styles.go
git commit -m "feat: add Lip Gloss theme with cyan accent dark palette"
```

---

### Task 5: KeyMap definitions

**Files:**
- Create: `internal/app/keys.go`

**Step 1: Create the keymap**

```go
// internal/app/keys.go
package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Add       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Status    key.Binding
	Subtask   key.Binding
	ToggleSub key.Binding
	Search    key.Binding
	Tab       key.Binding
	SortDate  key.Binding
	SortDue   key.Binding
	SortPrio  key.Binding
	Help      key.Binding
	Quit      key.Binding
	Enter     key.Binding
	Escape    key.Binding
}

var keys = keyMap{
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Status: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status"),
	),
	Subtask: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "subtask"),
	),
	ToggleSub: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "toggle sub"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "tab"),
	),
	SortDate: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("F1", "sort date"),
	),
	SortDue: key.NewBinding(
		key.WithKeys("f2"),
		key.WithHelp("F2", "sort due"),
	),
	SortPrio: key.NewBinding(
		key.WithKeys("f3"),
		key.WithHelp("F3", "sort prio"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Edit, k.Delete, k.Status, k.Subtask, k.Search, k.Tab, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Add, k.Edit, k.Delete},
		{k.Status, k.Subtask, k.ToggleSub},
		{k.Search, k.Tab, k.Help},
		{k.SortDate, k.SortDue, k.SortPrio},
		{k.Quit},
	}
}
```

**Step 2: Compile**

Run: `go build ./internal/app/...`
Expected: No errors.

**Step 3: Commit**

```bash
git add internal/app/keys.go
git commit -m "feat: add KeyMap with vim-style bindings and help integration"
```

---

### Task 6: Custom list item delegate

**Files:**
- Create: `internal/app/delegate.go`

**Step 1: Create the custom delegate**

```go
// internal/app/delegate.go
package app

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/task"
)

type taskDelegate struct{}

func newTaskDelegate() taskDelegate {
	return taskDelegate{}
}

func (d taskDelegate) Height() int  { return 2 }
func (d taskDelegate) Spacing() int { return 0 }

func (d taskDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	t, ok := item.(task.Task)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Status icon
	var statusIcon string
	switch t.Status {
	case task.InProgress:
		statusIcon = statusInProgressStyle.Render(t.Status.Icon())
	case task.Done:
		statusIcon = statusDoneStyle.Render(t.Status.Icon())
	default:
		statusIcon = statusPendingStyle.Render(t.Status.Icon())
	}

	// Priority label
	var prioLabel string
	switch t.Priority {
	case task.Urgent:
		prioLabel = priorityUrgentStyle.Render(t.Priority.Label())
	case task.High:
		prioLabel = priorityHighStyle.Render(t.Priority.Label())
	case task.Medium:
		prioLabel = priorityMedStyle.Render(t.Priority.Label())
	default:
		prioLabel = priorityLowStyle.Render(t.Priority.Label())
	}

	// Title line
	titleStyle := lipgloss.NewStyle().Foreground(white)
	if t.Status == task.Done {
		titleStyle = titleStyle.Strikethrough(true).Foreground(gray)
	}

	line1 := fmt.Sprintf(" %s %s %s", statusIcon, prioLabel, titleStyle.Render(t.Title))

	// Subtitle line: due date + subtask count
	var subtitle string
	if t.DueDate != nil {
		subtitle += fmt.Sprintf("due %s", t.DueDate.Format("Jan 02"))
	}
	if len(t.Subtasks) > 0 {
		done := 0
		for _, st := range t.Subtasks {
			if st.Completed {
				done++
			}
		}
		if subtitle != "" {
			subtitle += "  "
		}
		subtitle += fmt.Sprintf("[%d/%d]", done, len(t.Subtasks))
	}
	line2 := lipgloss.NewStyle().Foreground(gray).PaddingLeft(5).Render(subtitle)

	// Cursor / selection
	if isSelected {
		cursor := lipgloss.NewStyle().Foreground(cyan).Render("▸")
		line1 = cursor + line1[1:]
		line1 = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Render(line1)
		line2 = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Render(line2)
	}

	fmt.Fprintf(w, "%s\n%s", line1, line2)
}
```

**Step 2: Compile**

Run: `go build ./internal/app/...`
Expected: No errors.

**Step 3: Commit**

```bash
git add internal/app/delegate.go
git commit -m "feat: add custom list delegate with priority colors and status icons"
```

---

### Task 7: View rendering functions

**Files:**
- Create: `internal/ui/views.go`

**Step 1: Create the view helpers**

```go
// internal/ui/views.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/task"
)

var (
	cyan    = lipgloss.Color("#00BCD4")
	white   = lipgloss.Color("#FAFAFA")
	gray    = lipgloss.Color("#666666")
	dimGray = lipgloss.Color("#444444")
	green   = lipgloss.Color("#4CAF50")
	red     = lipgloss.Color("#F44336")
	yellow  = lipgloss.Color("#FFC107")
	magenta = lipgloss.Color("#E040FB")

	labelStyle = lipgloss.NewStyle().Foreground(gray).Width(12)
	valueStyle = lipgloss.NewStyle().Foreground(white)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(white)
)

// RenderTabs renders the tab bar.
func RenderTabs(activeTab int, allCount, activeCount, doneCount int, width int) string {
	tabs := []struct {
		label string
		count int
	}{
		{"All", allCount},
		{"Active", activeCount},
		{"Done", doneCount},
	}

	tabNormal := lipgloss.NewStyle().Padding(0, 2).Foreground(gray)
	tabActive := lipgloss.NewStyle().Padding(0, 2).Foreground(cyan).Bold(true).Reverse(true)

	var rendered []string
	appTitle := lipgloss.NewStyle().Bold(true).Foreground(cyan).Padding(0, 1).Render("Todo")
	rendered = append(rendered, appTitle)

	for i, t := range tabs {
		label := fmt.Sprintf("%s (%d)", t.label, t.count)
		if i == activeTab {
			rendered = append(rendered, tabActive.Render(label))
		} else {
			rendered = append(rendered, tabNormal.Render(label))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(dimGray).
		Width(width).
		Render(row)
}

// RenderDetail renders the task detail panel content.
func RenderDetail(t *task.Task, width int) string {
	if t == nil {
		return lipgloss.NewStyle().
			Foreground(gray).
			Align(lipgloss.Center).
			Width(width).
			Render("\n\n\nSelect a task to view details")
	}

	var sections []string

	// Title
	sections = append(sections, titleStyle.Render(t.Title))
	sections = append(sections, "")

	// Status
	statusStr := t.Status.Icon() + " " + t.Status.String()
	sections = append(sections, labelStyle.Render("Status")+statusStyle(t.Status).Render(statusStr))

	// Priority
	sections = append(sections, labelStyle.Render("Priority")+prioStyle(t.Priority).Render(t.Priority.String()))

	// Due date
	if t.DueDate != nil {
		sections = append(sections, labelStyle.Render("Due")+valueStyle.Render(t.DueDate.Format("Jan 02, 2006")))
	}

	// Created
	sections = append(sections, labelStyle.Render("Created")+valueStyle.Render(t.CreatedAt.Format("Jan 02, 2006")))

	// Tags
	if len(t.Tags) > 0 {
		tagStr := strings.Join(t.Tags, ", ")
		sections = append(sections, labelStyle.Render("Tags")+valueStyle.Render(tagStr))
	}

	// Description
	if t.Description != "" {
		sections = append(sections, "")
		sections = append(sections, labelStyle.Render("Description"))
		sections = append(sections, valueStyle.Render(t.Description))
	}

	// Subtasks
	if len(t.Subtasks) > 0 {
		sections = append(sections, "")
		doneCount := 0
		for _, st := range t.Subtasks {
			if st.Completed {
				doneCount++
			}
		}
		sections = append(sections, labelStyle.Render("Subtasks")+valueStyle.Render(fmt.Sprintf("%d/%d", doneCount, len(t.Subtasks))))
		sections = append(sections, renderProgressBar(doneCount, len(t.Subtasks), width-4))
		sections = append(sections, "")
		for _, st := range t.Subtasks {
			if st.Completed {
				sections = append(sections, lipgloss.NewStyle().Foreground(green).Render("  [x] "+st.Title))
			} else {
				sections = append(sections, lipgloss.NewStyle().Foreground(white).Render("  [ ] "+st.Title))
			}
		}
	}

	return strings.Join(sections, "\n")
}

func renderProgressBar(done, total, width int) string {
	if total == 0 || width < 4 {
		return ""
	}
	barWidth := width - 2
	if barWidth > 40 {
		barWidth = 40
	}
	filled := barWidth * done / total
	empty := barWidth - filled

	bar := lipgloss.NewStyle().Foreground(cyan).Render(strings.Repeat("█", filled))
	bar += lipgloss.NewStyle().Foreground(dimGray).Render(strings.Repeat("░", empty))
	return "  " + bar
}

// RenderStatusBar renders the bottom status bar.
func RenderStatusBar(total, done, inProgress int, width int) string {
	keyStyle := lipgloss.NewStyle().Foreground(cyan)
	dimStyle := lipgloss.NewStyle().Foreground(gray)

	left := dimStyle.Render(fmt.Sprintf(" %d tasks | %d done | %d active", total, done, inProgress))

	bindings := []struct{ key, desc string }{
		{"a", "add"}, {"e", "edit"}, {"d", "del"}, {"s", "status"},
		{"t", "sub"}, {"/", "find"}, {"?", "help"},
	}
	var parts []string
	for _, b := range bindings {
		parts = append(parts, keyStyle.Render(b.key)+dimStyle.Render(":"+b.desc))
	}
	right := strings.Join(parts, dimStyle.Render(" "))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// RenderConfirmDialog renders a yes/no confirmation dialog.
func RenderConfirmDialog(title, message string, width, height int) string {
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(white).
		Render(title) + "\n\n" +
		lipgloss.NewStyle().Foreground(gray).Render(message) + "\n\n" +
		lipgloss.NewStyle().Foreground(gray).Render("[y] confirm  [n/esc] cancel")

	dialog := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(red).
		Padding(1, 2).
		Width(50).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}

func statusStyle(s task.Status) lipgloss.Style {
	switch s {
	case task.InProgress:
		return lipgloss.NewStyle().Foreground(yellow)
	case task.Done:
		return lipgloss.NewStyle().Foreground(green)
	default:
		return lipgloss.NewStyle().Foreground(gray)
	}
}

func prioStyle(p task.Priority) lipgloss.Style {
	switch p {
	case task.Urgent:
		return lipgloss.NewStyle().Foreground(magenta)
	case task.High:
		return lipgloss.NewStyle().Foreground(red)
	case task.Medium:
		return lipgloss.NewStyle().Foreground(yellow)
	default:
		return lipgloss.NewStyle().Foreground(green)
	}
}
```

**Step 2: Compile**

Run: `go build ./internal/ui/...`
Expected: No errors.

**Step 3: Commit**

```bash
git add internal/ui/views.go
git commit -m "feat: add view rendering - tabs, detail panel, status bar, confirm dialog"
```

---

### Task 8: Huh form builders

**Files:**
- Create: `internal/ui/form.go`

**Step 1: Create form builders**

```go
// internal/ui/form.go
package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/roniel/todo-app/internal/task"
)

// TaskFormData holds the form field values.
type TaskFormData struct {
	Title       string
	Description string
	Priority    task.Priority
	DueDate     string
	Tags        string
}

// NewTaskForm creates a Huh form for adding a new task.
func NewTaskForm(data *TaskFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&data.Title).
				Validate(huh.ValidateNotEmpty()),

			huh.NewText().
				Title("Description").
				Value(&data.Description).
				Lines(3),

			huh.NewSelect[task.Priority]().
				Title("Priority").
				Options(
					huh.NewOption("Low", task.Low),
					huh.NewOption("Medium", task.Medium).Selected(true),
					huh.NewOption("High", task.High),
					huh.NewOption("Urgent", task.Urgent),
				).
				Value(&data.Priority),

			huh.NewInput().
				Title("Due Date").
				Placeholder("YYYY-MM-DD").
				Value(&data.DueDate).
				Validate(validateOptionalDate),

			huh.NewInput().
				Title("Tags").
				Placeholder("comma separated").
				Value(&data.Tags),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(false)
}

// EditTaskForm creates a Huh form for editing an existing task.
func EditTaskForm(data *TaskFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&data.Title).
				Validate(huh.ValidateNotEmpty()),

			huh.NewText().
				Title("Description").
				Value(&data.Description).
				Lines(3),

			huh.NewSelect[task.Priority]().
				Title("Priority").
				Options(
					huh.NewOption("Low", task.Low).Selected(data.Priority == task.Low),
					huh.NewOption("Medium", task.Medium).Selected(data.Priority == task.Medium),
					huh.NewOption("High", task.High).Selected(data.Priority == task.High),
					huh.NewOption("Urgent", task.Urgent).Selected(data.Priority == task.Urgent),
				).
				Value(&data.Priority),

			huh.NewInput().
				Title("Due Date").
				Placeholder("YYYY-MM-DD").
				Value(&data.DueDate).
				Validate(validateOptionalDate),

			huh.NewInput().
				Title("Tags").
				Placeholder("comma separated").
				Value(&data.Tags),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(false)
}

// SubtaskForm creates a simple single-field form for adding a subtask.
func SubtaskForm(title *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Subtask").
				Value(title).
				Validate(huh.ValidateNotEmpty()),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(false)
}

func validateOptionalDate(s string) error {
	if s == "" {
		return nil
	}
	_, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return fmt.Errorf("use YYYY-MM-DD format")
	}
	return nil
}
```

**Step 2: Compile**

Run: `go build ./internal/ui/...`
Expected: No errors.

**Step 3: Commit**

```bash
git add internal/ui/form.go
git commit -m "feat: add Huh form builders for task add/edit and subtask"
```

---

### Task 9: Main app model - Bubbletea Model, Update, View

**Files:**
- Create: `internal/app/model.go`
- Modify: `cmd/todo/main.go`

This is the largest task. It wires everything together.

**Step 1: Create the app model**

```go
// internal/app/model.go
package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
)

type mode int

const (
	modeNormal mode = iota
	modeAdd
	modeEdit
	modeConfirmDelete
	modeSubtask
	modeHelp
)

type sortOrder int

const (
	sortCreated sortOrder = iota
	sortDue
	sortPriority
)

type Model struct {
	store    *task.Store
	tasks    []task.Task
	list     list.Model
	viewport viewport.Model
	help     help.Model
	form     *huh.Form
	formData *ui.TaskFormData
	subtaskTitle string

	mode      mode
	activeTab int // 0=All, 1=Active, 2=Done
	sortBy    sortOrder
	width     int
	height    int
	ready     bool
	err       error
}

// tasksLoaded is a custom message for initial data load.
type tasksLoaded struct {
	tasks []task.Task
	err   error
}

func New(store *task.Store) Model {
	delegate := newTaskDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(false)
	l.DisableQuitKeybindings()

	vp := viewport.New(0, 0)
	h := help.New()

	return Model{
		store:    store,
		list:     l,
		viewport: vp,
		help:     h,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.store.List()
		return tasksLoaded{tasks: tasks, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoaded:
		m.tasks = msg.tasks
		if msg.err != nil {
			m.err = msg.err
		}
		m.refreshList()
		m.updateDetail()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeComponents()
		return m, nil

	case tea.KeyMsg:
		// If in form mode, delegate to form
		if m.mode == modeAdd || m.mode == modeEdit {
			return m.updateForm(msg)
		}
		if m.mode == modeSubtask {
			return m.updateSubtaskForm(msg)
		}
		if m.mode == modeConfirmDelete {
			return m.updateConfirmDelete(msg)
		}
		if m.mode == modeHelp {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
				m.mode = modeNormal
				return m, nil
			}
			return m, nil
		}

		// Normal mode keybindings
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.mode = modeHelp
			return m, nil

		case key.Matches(msg, keys.Add):
			m.formData = &ui.TaskFormData{Priority: task.Medium}
			m.form = ui.NewTaskForm(m.formData)
			m.mode = modeAdd
			return m, m.form.Init()

		case key.Matches(msg, keys.Edit):
			selected := m.selectedTask()
			if selected == nil {
				return m, nil
			}
			dueStr := ""
			if selected.DueDate != nil {
				dueStr = selected.DueDate.Format(time.DateOnly)
			}
			m.formData = &ui.TaskFormData{
				Title:       selected.Title,
				Description: selected.Description,
				Priority:    selected.Priority,
				DueDate:     dueStr,
				Tags:        strings.Join(selected.Tags, ", "),
			}
			m.form = ui.EditTaskForm(m.formData)
			m.mode = modeEdit
			return m, m.form.Init()

		case key.Matches(msg, keys.Delete):
			if m.selectedTask() != nil {
				m.mode = modeConfirmDelete
			}
			return m, nil

		case key.Matches(msg, keys.Status):
			selected := m.selectedTask()
			if selected != nil {
				selected.Status = selected.Status.Next()
				m.store.Update(selected)
				m.reload()
			}
			return m, nil

		case key.Matches(msg, keys.Subtask):
			if m.selectedTask() != nil {
				m.subtaskTitle = ""
				m.form = ui.SubtaskForm(&m.subtaskTitle)
				m.mode = modeSubtask
				return m, m.form.Init()
			}
			return m, nil

		case key.Matches(msg, keys.ToggleSub):
			selected := m.selectedTask()
			if selected != nil {
				for _, st := range selected.Subtasks {
					if !st.Completed {
						m.store.ToggleSubtask(st.ID)
						m.reload()
						break
					}
				}
			}
			return m, nil

		case key.Matches(msg, keys.Tab):
			m.activeTab = (m.activeTab + 1) % 3
			m.refreshList()
			m.updateDetail()
			return m, nil

		case key.Matches(msg, keys.SortDate):
			m.sortBy = sortCreated
			m.sortTasks()
			return m, nil

		case key.Matches(msg, keys.SortDue):
			m.sortBy = sortDue
			m.sortTasks()
			return m, nil

		case key.Matches(msg, keys.SortPrio):
			m.sortBy = sortPriority
			m.sortTasks()
			return m, nil

		case key.Matches(msg, keys.Search):
			// Enable built-in list filtering
			m.list.SetShowFilter(true)
			m.list.SetFilteringEnabled(true)
			// Trigger filter mode by sending "/" to the list
		}
	}

	// Delegate to list for navigation and filtering
	var cmd tea.Cmd
	prevIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIndex {
		m.updateDetail()
	}

	return m, cmd
}

func (m *Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		m.formData = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			m.submitTaskForm()
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) updateSubtaskForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			selected := m.selectedTask()
			if selected != nil && m.subtaskTitle != "" {
				m.store.AddSubtask(selected.ID, m.subtaskTitle)
			}
			m.mode = modeNormal
			m.form = nil
			m.reload()
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		selected := m.selectedTask()
		if selected != nil {
			m.store.Delete(selected.ID)
			m.reload()
		}
		m.mode = modeNormal
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) submitTaskForm() {
	if m.formData == nil {
		return
	}

	var dueDate *time.Time
	if m.formData.DueDate != "" {
		if d, err := time.Parse(time.DateOnly, m.formData.DueDate); err == nil {
			dueDate = &d
		}
	}

	var tags []string
	if m.formData.Tags != "" {
		for _, t := range strings.Split(m.formData.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	if m.mode == modeAdd {
		t := &task.Task{
			Title:       m.formData.Title,
			Description: m.formData.Description,
			Priority:    m.formData.Priority,
			DueDate:     dueDate,
			Tags:        tags,
		}
		m.store.Create(t)
	} else if m.mode == modeEdit {
		selected := m.selectedTask()
		if selected != nil {
			selected.Title = m.formData.Title
			selected.Description = m.formData.Description
			selected.Priority = m.formData.Priority
			selected.DueDate = dueDate
			selected.Tags = tags
			m.store.Update(selected)
		}
	}

	m.mode = modeNormal
	m.form = nil
	m.formData = nil
	m.reload()
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	// Calculate counts
	allCount := len(m.tasks)
	var activeCount, doneCount int
	for _, t := range m.tasks {
		switch t.Status {
		case task.Done:
			doneCount++
		default:
			activeCount++
		}
	}
	inProgress := 0
	for _, t := range m.tasks {
		if t.Status == task.InProgress {
			inProgress++
		}
	}

	// Header tabs
	header := ui.RenderTabs(m.activeTab, allCount, activeCount, doneCount, m.width)

	// Content area height
	contentHeight := m.height - lipgloss.Height(header) - 1 // 1 for status bar

	// List panel (40% width)
	listWidth := m.width * 2 / 5
	detailWidth := m.width - listWidth

	listPanel := listPanelFocusedStyle.
		Width(listWidth - 2). // account for border
		Height(contentHeight - 2).
		Render(m.list.View())

	// Detail panel
	detailContent := m.viewport.View()
	detailPanel := detailPanelStyle.
		Width(detailWidth - 2).
		Height(contentHeight - 2).
		Render(detailContent)

	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)

	// Status bar
	statusBar := ui.RenderStatusBar(allCount, doneCount, inProgress, m.width)

	// Combine
	view := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)

	// Overlay dialogs
	switch m.mode {
	case modeAdd, modeEdit:
		title := "New Task"
		if m.mode == modeEdit {
			title = "Edit Task"
		}
		formView := m.form.View()
		dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render(title) + "\n\n" + formView
		dialog := dialogStyle.Render(dialogContent)
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeSubtask:
		formView := m.form.View()
		dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render("Add Subtask") + "\n\n" + formView
		dialog := dialogStyle.Render(dialogContent)
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeConfirmDelete:
		selected := m.selectedTask()
		title := ""
		if selected != nil {
			title = selected.Title
		}
		view = ui.RenderConfirmDialog("Delete Task?", fmt.Sprintf("Delete \"%s\"?", title), m.width, m.height)

	case modeHelp:
		helpView := m.renderHelpOverlay()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
	}

	return view
}

func (m *Model) resizeComponents() {
	headerHeight := 3
	statusBarHeight := 1
	contentHeight := m.height - headerHeight - statusBarHeight

	listWidth := m.width*2/5 - 4 // borders + padding
	detailWidth := m.width - m.width*2/5 - 6

	m.list.SetSize(listWidth, contentHeight-2)
	m.viewport.Width = detailWidth
	m.viewport.Height = contentHeight - 4
	m.help.Width = m.width
}

func (m *Model) selectedTask() *task.Task {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	t, ok := item.(task.Task)
	if !ok {
		return nil
	}
	// Find the original task in our slice (so changes persist)
	for i := range m.tasks {
		if m.tasks[i].ID == t.ID {
			return &m.tasks[i]
		}
	}
	return nil
}

func (m *Model) reload() {
	tasks, err := m.store.List()
	if err != nil {
		m.err = err
		return
	}
	m.tasks = tasks
	m.sortTasks()
	m.refreshList()
	m.updateDetail()
}

func (m *Model) refreshList() {
	filtered := m.filteredTasks()
	items := make([]list.Item, len(filtered))
	for i, t := range filtered {
		items[i] = t
	}
	m.list.SetItems(items)
}

func (m *Model) filteredTasks() []task.Task {
	if m.activeTab == 0 {
		return m.tasks
	}
	var filtered []task.Task
	for _, t := range m.tasks {
		switch m.activeTab {
		case 1: // Active (Pending + InProgress)
			if t.Status != task.Done {
				filtered = append(filtered, t)
			}
		case 2: // Done
			if t.Status == task.Done {
				filtered = append(filtered, t)
			}
		}
	}
	return filtered
}

func (m *Model) updateDetail() {
	selected := m.selectedTask()
	detailWidth := m.width - m.width*2/5 - 6
	content := ui.RenderDetail(selected, detailWidth)
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m *Model) sortTasks() {
	switch m.sortBy {
	case sortDue:
		sort.Slice(m.tasks, func(i, j int) bool {
			if m.tasks[i].DueDate == nil && m.tasks[j].DueDate == nil {
				return m.tasks[i].CreatedAt.After(m.tasks[j].CreatedAt)
			}
			if m.tasks[i].DueDate == nil {
				return false
			}
			if m.tasks[j].DueDate == nil {
				return true
			}
			return m.tasks[i].DueDate.Before(*m.tasks[j].DueDate)
		})
	case sortPriority:
		sort.Slice(m.tasks, func(i, j int) bool {
			return m.tasks[i].Priority > m.tasks[j].Priority
		})
	default: // sortCreated
		sort.Slice(m.tasks, func(i, j int) bool {
			return m.tasks[i].CreatedAt.After(m.tasks[j].CreatedAt)
		})
	}
	m.refreshList()
	m.updateDetail()
}

func (m Model) renderHelpOverlay() string {
	helpLines := []struct{ key, desc string }{
		{"j/k ↑/↓", "Navigate tasks"},
		{"Enter", "Select task"},
		{"a", "Add new task"},
		{"e", "Edit selected task"},
		{"d", "Delete selected task"},
		{"s", "Cycle task status"},
		{"t", "Add subtask"},
		{"x", "Toggle next subtask"},
		{"/", "Search / filter"},
		{"Tab", "Switch tab"},
		{"F1", "Sort by created date"},
		{"F2", "Sort by due date"},
		{"F3", "Sort by priority"},
		{"?", "Toggle this help"},
		{"q", "Quit"},
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Keyboard Shortcuts"))
	lines = append(lines, "")
	for _, h := range helpLines {
		k := lipgloss.NewStyle().Foreground(cyan).Width(14).Render(h.key)
		d := lipgloss.NewStyle().Foreground(gray).Render(h.desc)
		lines = append(lines, k+d)
	}
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc or ? to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 3).
		Width(45).
		Render(content)
}
```

**Step 2: Update main.go to launch the Bubbletea app**

```go
// cmd/todo/main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/roniel/todo-app/internal/app"
	"github.com/roniel/todo-app/internal/task"
)

func main() {
	store, err := task.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	m := app.New(store)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Compile**

Run: `go build ./cmd/todo`
Expected: No errors, `todo` binary created.

**Step 4: Run it**

Run: `./todo`
Expected: App launches in alternate screen with empty task list, tabs, status bar. Press `q` to quit.

**Step 5: Commit**

```bash
git add internal/app/model.go cmd/todo/main.go
git commit -m "feat: wire up main Bubbletea app with list, viewport, forms, keybindings"
```

---

### Task 10: Integration testing - end-to-end smoke test

**Files:**
- No new files

**Step 1: Build the binary**

Run: `go build -o todo-app ./cmd/todo`
Expected: Binary `todo-app` created without errors.

**Step 2: Run go vet**

Run: `go vet ./...`
Expected: No issues.

**Step 3: Smoke test in terminal**

Run: `./todo-app`
Verify:
- App launches in alternate screen
- Tab bar shows "All (0) | Active (0) | Done (0)"
- Status bar shows keybinding hints
- Press `a` -> Huh form appears with Title, Description, Priority, Due Date, Tags
- Fill in a task title, press Enter through fields -> task appears in list
- Press `e` -> form opens pre-filled with task data
- Press `s` -> task status cycles Pending -> InProgress -> Done
- Press `Tab` -> switches between All/Active/Done tabs
- Press `d` -> confirmation dialog appears, `y` deletes, `n` cancels
- Press `t` -> subtask form appears
- Press `x` -> toggles subtask
- Press `?` -> help overlay appears
- Press `q` -> app exits cleanly

**Step 4: Verify data persists**

Run: `./todo-app` again
Expected: Previously created tasks still appear (SQLite persistence).

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: integration verified - full TUI working with SQLite persistence"
```

---

### Task 11: Update project documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `.gitignore`

**Step 1: Update .gitignore for Go**

Replace contents of `.gitignore` with:

```
# Go
todo-app
/todo
*.exe
*.test
*.out
vendor/

# IDE
.idea/
.vscode/
*.swp

# Data
*.db

# OS
.DS_Store
```

**Step 2: Update CLAUDE.md for Go project**

Update the CLAUDE.md to reflect the new Go tech stack:
- Replace Java/Maven/TamboUI references with Go/Bubbletea/Bubbles/Lip Gloss/Huh/SQLite
- Update build commands to `go build`, `go test`, `go vet`
- Update project structure section
- Keep workflow orchestration and core principles sections as-is

**Step 3: Commit**

```bash
git add .gitignore CLAUDE.md
git commit -m "docs: update project docs for Go + Bubbletea stack"
```

---

### Task 12: Final cleanup and go mod tidy

**Files:**
- No new files

**Step 1: Clean up unused dependencies**

Run: `go mod tidy`
Expected: go.mod and go.sum cleaned up.

**Step 2: Verify everything builds clean**

Run: `go build ./... && go vet ./...`
Expected: No errors.

**Step 3: Run final build**

Run: `go build -o todo-app ./cmd/todo`
Expected: Single binary `todo-app` created.

**Step 4: Final commit**

```bash
git add go.mod go.sum
git commit -m "chore: final go mod tidy and clean build"
```
