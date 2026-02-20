# Todo App - Project Guide

## Project Overview

A modern terminal user interface (TUI) task management application built with **Go** and the **Charm** ecosystem.

### Tech Stack
- **Language**: Go 1.23+
- **TUI Framework**: Bubbletea v1.3.x (MVU pattern)
- **Components**: Bubbles v1.0.x (list, viewport, help, key, textinput)
- **Styling**: Lip Gloss v1.1.x
- **Forms**: Huh v0.8.x (task add/edit dialogs)
- **Database**: SQLite via modernc.org/sqlite (CGO-free)

### Key Dependencies (`go.mod`)
```
github.com/charmbracelet/bubbletea v1.3.10
github.com/charmbracelet/bubbles v1.0.0
github.com/charmbracelet/lipgloss v1.1.0
github.com/charmbracelet/huh v0.8.0
modernc.org/sqlite v1.46.1
```

---

## Application Features

### Core Functionality
- **Task Management**: Create, view, edit, and delete tasks
- **Subtask Support**: Tasks can have subtasks with independent completion state
- **Status Tracking**: Cycle tasks between Pending, In Progress, Done
- **Tab Navigation**: All / Active / Done tabs with counts
- **Task Details**: Right panel shows description, subtasks, progress bar
- **Date Tracking**: Automatic creation date + optional due date
- **Sorting**: Sort by creation date (F1), due date (F2), or priority (F3)
- **Search**: Fuzzy search/filter via built-in bubbles list filtering
- **Persistence**: SQLite database at `~/.todo-app/todo.db`

### Additional Features
- Priority levels (Low, Medium, High, Urgent) with color coding
- Tags (comma-separated)
- Keyboard-driven navigation (vim-style j/k + arrows)
- Status bar with task counts and keybinding hints
- Confirmation dialogs for destructive actions
- Huh forms with validation for task create/edit
- Custom list delegate with status icons, priority labels, subtask counts

### UI Layout
```
┌──────────────────────────────────────────────────────────────────┐
│  Todo  │  All (7)  │  Active (4)  │  Done (3)                   │
├────────────────────────┬─────────────────────────────────────────┤
│  Task list (bubbles)   │  Task detail (viewport)                 │
│  - Custom delegate     │  - Status, Priority, Due, Tags          │
│  - Fuzzy search        │  - Description                          │
│  - Colored items       │  - Subtasks + progress bar              │
├────────────────────────┴─────────────────────────────────────────┤
│  a:add  e:edit  d:del  s:status  t:sub  /:find  ?:help          │
└──────────────────────────────────────────────────────────────────┘
```

---

## Architecture

### Bubbletea MVU (Model-Update-View)
- **Model**: `internal/app/model.go` - main state struct with list, viewport, form, mode tracking
- **Update**: Handles tea.Msg dispatch - key events route to mode-specific handlers
- **View**: Renders layout with header tabs, split panels, status bar, and modal overlays

### Project Structure
```
cmd/todo/main.go              # Entry point
internal/
  app/
    model.go                  # Main Bubbletea Model + Update + View
    keys.go                   # KeyMap definitions (key.Binding)
    styles.go                 # Lip Gloss styles (cyan accent dark theme)
    delegate.go               # Custom list.ItemDelegate for task rendering
  task/
    task.go                   # Domain model (Task, Subtask, Status, Priority)
    store.go                  # SQLite repository (CRUD, subtasks, tags)
  ui/
    views.go                  # View rendering (tabs, detail, status bar, dialogs)
    form.go                   # Huh form builders (task add/edit, subtask)
go.mod / go.sum
```

### Data Model
```go
type Task struct {
    ID          int64
    Title       string
    Description string
    Status      Status      // Pending, InProgress, Done
    Priority    Priority    // Low, Medium, High, Urgent
    DueDate     *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Subtasks    []Subtask
    Tags        []string
}

type Subtask struct {
    ID        int64
    Title     string
    Completed bool
    Position  int
}
```

### SQLite Schema
Database at `~/.todo-app/todo.db` with tables: `tasks`, `subtasks`, `tags` (with ON DELETE CASCADE).

---

## Key Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Navigate task list |
| `a` | Add new task (Huh form) |
| `e` | Edit selected task |
| `d` | Delete selected task (with confirmation) |
| `s` | Cycle task status |
| `t` | Add subtask |
| `x` | Toggle next incomplete subtask |
| `/` | Search / filter |
| `Tab` | Switch tabs (All/Active/Done) |
| `F1`/`F2`/`F3` | Sort by created/due/priority |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

---

## Workflow Orchestration

### 1. Plan Mode Default
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately
- Write detailed specs upfront to reduce ambiguity

### 2. Subagent Strategy
- Use subagents liberally to keep main context window clean
- Offload research, exploration, and parallel analysis to subagents
- One task per subagent for focused execution

### 3. Verification Before Done
- Never mark a task complete without proving it works
- Run `go build`, `go vet`, `go test` before claiming done
- Ask yourself: "Would a staff engineer approve this?"

### 4. Autonomous Bug Fixing
- When given a bug report: just fix it
- Point at logs, errors, failing tests - then resolve them

---

## Core Principles

- **Simplicity First**: Make every change as simple as possible
- **No Laziness**: Find root causes. No temporary fixes
- **Minimal Impact**: Changes should only touch what's necessary
- **Test Everything**: Verify behavior manually and with tests
- **Consistent Style**: Follow Go conventions, keep packages focused
- **No AI Attribution**: Never reference Claude, AI, or any assistant in commit messages, code comments, or any project file

---

## Build & Run

```bash
# Build
go build ./cmd/todo

# Build named binary
go build -o todo-app ./cmd/todo

# Run directly
go run ./cmd/todo

# Run tests
go test ./...

# Vet
go vet ./...

# Tidy dependencies
go mod tidy
```

---

## Charm Ecosystem Links
- Bubbletea: https://github.com/charmbracelet/bubbletea
- Bubbles: https://github.com/charmbracelet/bubbles
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Huh: https://github.com/charmbracelet/huh
