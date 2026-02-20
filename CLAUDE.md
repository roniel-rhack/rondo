# RonDO - Project Guide

## Project Overview

**RonDO** is a modern terminal user interface (TUI) productivity app built with **Go** and the **Charm** ecosystem. It combines task management with a daily journal in a single keyboard-driven interface.

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

### Task Management
- **CRUD**: Create, view, edit, and delete tasks
- **Subtask Support**: Tasks can have subtasks with independent completion state
- **Status Tracking**: Cycle tasks between Pending, In Progress, Done
- **Tab Navigation**: All / Active / Done tabs with counts
- **Task Details**: Right panel shows description, subtasks, progress bar
- **Date Tracking**: Automatic creation date + optional due date
- **Sorting**: Sort by creation date (F1), due date (F2), or priority (F3)
- **Search**: Fuzzy search/filter via built-in bubbles list filtering
- **Priority Levels**: Low, Medium, High, Urgent with color coding
- **Tags**: Comma-separated tag support

### Journal
- **Daily Notes**: One note per calendar day, auto-created
- **Entries**: Multiple timestamped entries per note
- **Edit/Delete Entries**: Cursor-based entry selection with edit and delete
- **Hide/Restore Notes**: Hide old notes, toggle visibility with `H`
- **Smart Date Labels**: "Today", "Yesterday", weekday names, or full dates
- **Search Notes**: Filter notes by date

### General
- Keyboard-driven navigation (vim-style j/k + arrows)
- Two-panel layout with focus switching (1/2 keys)
- Status bar with context-sensitive keybinding hints
- Confirmation dialogs for all destructive actions
- Huh forms with validation for all input
- Dark theme with cyan accent colors
- Persistence via SQLite at `~/.todo-app/todo.db`

### UI Layout
```
┌──────────────────────────────────────────────────────────────────┐
│  RonDO  │  All (7)  │  Active (4)  │  Done (3)  │  Journal (5) │
├────────────────────────┬─────────────────────────────────────────┤
│  1: Panel (list)       │  2: Panel (detail/viewport)             │
│  - Custom delegate     │  - Context-sensitive content             │
│  - Fuzzy search        │  - Cursor selection in both panels       │
│  - Colored items       │  - Subtasks/entries with progress        │
├────────────────────────┴─────────────────────────────────────────┤
│  Context-sensitive status bar with keybinding hints              │
└──────────────────────────────────────────────────────────────────┘
```

---

## Architecture

### Bubbletea MVU (Model-Update-View)
- **Model**: `internal/app/model.go` — main state struct with list, viewport, form, mode tracking
- **Update**: Global keys (Quit/Help/Tab) handled first, then per-tab dispatch
- **View**: Renders layout with header tabs, split panels, status bar, and modal overlays

### Project Structure
```
cmd/todo/main.go                # Entry point
internal/
  app/
    model.go                    # Main Bubbletea Model + Update + View
    model_journal.go            # Journal tab handlers (update, form, confirm, view)
    keys.go                     # KeyMap definitions (key.Binding)
    styles.go                   # Lip Gloss styles (cyan accent dark theme)
    delegate.go                 # Custom list.ItemDelegate for task rendering
    delegate_journal.go         # Custom list.ItemDelegate for journal notes
  database/
    db.go                       # SQLite connection (WAL mode, foreign keys)
  journal/
    journal.go                  # Domain model (Note, Entry, DateTitle)
    store.go                    # SQLite repository (CRUD, batch queries, transactions)
  task/
    task.go                     # Domain model (Task, Subtask, Status, Priority)
    store.go                    # SQLite repository (CRUD, subtasks, tags)
  ui/
    colors.go                   # Shared color palette
    views.go                    # View rendering (tabs, detail, status bar, dialogs)
    form.go                     # Huh form builders + form data types
go.mod / go.sum
```

### Data Model
```go
// Task management
type Task struct {
    ID, Title, Description, Status, Priority
    DueDate, CreatedAt, UpdatedAt
    Subtasks []Subtask
    Tags     []string
}

type Subtask struct {
    ID, Title, Completed, Position
}

// Journal
type Note struct {
    ID, Date, Hidden, CreatedAt, UpdatedAt
    Entries []Entry
}

type Entry struct {
    ID, NoteID, Body, CreatedAt
}
```

### Mode System
The app uses a mode enum to track UI state:
- `modeNormal` — default navigation
- `modeAdd`, `modeEdit` — task forms
- `modeSubtask`, `modeEditSubtask` — subtask forms
- `modeConfirmDelete`, `modeConfirmDeleteSubtask` — task/subtask delete confirmation
- `modeJournalAdd`, `modeJournalEdit` — journal entry forms
- `modeJournalConfirmHide`, `modeJournalConfirmDelete` — journal confirmations
- `modeHelp` — help overlay

### Key Patterns
- **Global keys** (Quit/Help/Tab) are handled once in `Update()` before per-tab dispatch
- **`switchTab()`** centralizes tab switching logic
- **`reload()` / `reloadJournal()`** return errors, callers handle them
- **Cursor tracking**: `subtaskIdx` for task detail, `entryIdx` for journal entries
- **Confirm dialogs**: `ui.RenderConfirmDialogBox` with variadic border color
- **Batch queries**: Journal entries loaded in a single query (no N+1)
- **Transactions**: All multi-statement writes use `db.Begin()`/`tx.Commit()`
- **Date parsing**: `time.ParseInLocation` for date-only fields (timezone-correct)

### SQLite Schema
Database at `~/.todo-app/todo.db` with tables:
- `tasks`, `subtasks`, `tags` — task management (ON DELETE CASCADE)
- `journal_notes`, `journal_entries` — journal (ON DELETE CASCADE, indexed)

---

## Keyboard Shortcuts

### Global
| Key | Action |
|-----|--------|
| `Tab` | Switch tabs (All → Active → Done → Journal) |
| `1` / `2` | Focus left panel / right panel |
| `Esc` | Return to left panel / clear filter |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### Task Tabs — Panel 1 (Tasks)
| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Navigate task list |
| `a` | Add new task |
| `e` | Edit selected task |
| `d` | Delete selected task |
| `s` | Cycle task status |
| `t` | Add subtask |
| `/` | Search / filter |
| `F1`/`F2`/`F3` | Sort by created/due/priority |

### Task Tabs — Panel 2 (Details)
| Key | Action |
|-----|--------|
| `j`/`k` | Navigate subtasks |
| `a` | Add subtask |
| `e` | Edit subtask |
| `d` | Delete subtask |
| `s` | Toggle subtask completion |

### Journal Tab — Panel 1 (Notes)
| Key | Action |
|-----|--------|
| `j`/`k` | Navigate notes |
| `a` | Add entry to today's note |
| `h` | Hide / restore selected note |
| `H` | Toggle show hidden notes |
| `/` | Search / filter notes |

### Journal Tab — Panel 2 (Entries)
| Key | Action |
|-----|--------|
| `j`/`k` | Navigate entries (cursor) |
| `a` | Add entry to today's note |
| `e` | Edit selected entry |
| `d` | Delete selected entry |

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
- Point at logs, errors, failing tests — then resolve them

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
go build -o rondo ./cmd/todo

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
