# RonDO - Project Guide

## Project Overview

**RonDO** is a modern terminal user interface (TUI) productivity app built with **Go** and the **Charm** ecosystem. It combines task management with a daily journal in a single keyboard-driven interface.

### Tech Stack
- **Language**: Go 1.23+
- **TUI Framework**: Bubbletea v1.3.x (MVU pattern)
- **Components**: Bubbles v1.0.x (list, viewport, help, key, textinput)
- **Styling**: Lip Gloss v1.1.x (including lipgloss/table for styled CLI tables)
- **Forms**: Huh v0.8.x (task add/edit dialogs)
- **CLI Framework**: Cobra v1.10.x (subcommands, flags, completions)
- **Database**: SQLite via modernc.org/sqlite (CGO-free)
- **Terminal Detection**: charmbracelet/x/term (width), go-isatty (TTY detection)

### Key Dependencies (`go.mod`)
```
github.com/charmbracelet/bubbletea v1.3.10
github.com/charmbracelet/bubbles v1.0.0
github.com/charmbracelet/lipgloss v1.1.0
github.com/charmbracelet/huh v0.8.0
github.com/charmbracelet/x/term v0.2.2
github.com/spf13/cobra v1.10.2
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
- **Recurring Tasks**: Daily, weekly, monthly, or yearly recurrence; auto-spawns next occurrence on completion
- **Task Dependencies**: Block tasks by other task IDs; CLI via `--blocks` flag on `add`/`edit`, `--clear-blocks` on `edit`
- **Delete Guard**: Refuses to delete tasks that block others; CLI returns exit code 1 unless `--cascade`; TUI shows yellow warning dialog requiring double-confirm
- **Self-Block Guard**: Store-level rejection of self-referential dependencies
- **Metadata**: Structured key-value pairs (`--meta key=value`) stored as JSON; filterable with AND logic via `rondo list --meta key=value`
- **Time Logging**: Log time spent on tasks with optional notes

### Journal
- **Daily Notes**: One note per calendar day, auto-created
- **Entries**: Multiple timestamped entries per note
- **Edit/Delete Entries**: Cursor-based entry selection with edit and delete
- **Hide/Restore Notes**: Hide old notes, toggle visibility with `H`
- **Smart Date Labels**: "Today", "Yesterday", weekday names, or full dates
- **Search Notes**: Filter notes by date

### Pomodoro Timer
- **Full Pomodoro cycle**: Work → Short Break → Work → ... → Long Break (4-session sets)
- **Session types**: Work (🍅), Short Break (☕), Long Break (🌿) with distinct colors
- **Cycle indicator**: ●●●○ showing progress through 4-session set
- **Configurable**: Durations, daily goal, auto-start breaks via `P` settings form or config.json
- **Notifications**: Terminal bell on phase completion
- **Stats**: Daily goal progress, weekly summary, streak tracking in `G` overlay
- **Task linkage**: Focus sessions linked to selected task

### General
- Keyboard-driven navigation (vim-style j/k + arrows)
- Two-panel layout with focus switching (1/2 keys), resizable with `<`/`>`
- Status bar with context-sensitive keybinding hints
- Confirmation dialogs for all destructive actions
- Huh forms with validation for all input
- Adaptive color theme (auto-detects light/dark terminal)
- Persistence via SQLite at `~/.todo-app/todo.db`
- Auto backups at `~/.todo-app/backups/`
- Config file at `~/.todo-app/config.json`
- Undo last destructive action (`Ctrl+Z`)

### CLI Mode
Full Cobra-based CLI with all features available as subcommands:
- **Global flags**: `--format table|json`, `--json`, `--quiet`, `--no-color`
- **TTY-aware output**: Styled tables with Unicode borders + ANSI colors when TTY; plain tabwriter when piped
- **Commands**: `add`, `done`, `list`, `show`, `edit`, `delete`, `status`, `subtask`, `timelog`, `note`, `recur`, `journal`, `focus`, `export`, `stats`, `config`, `batch`, `completion`
- **Shell completions**: bash, zsh, fish, powershell via `rondo completion`

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
cmd/todo/main.go                # Entry point (TUI + CLI dispatch)
internal/
  app/
    model.go                    # Main Bubbletea Model + Update + View
    model_journal.go            # Journal tab handlers
    model_forms.go              # Form submission + confirmation dialogs
    model_overlays.go           # Help, stats, blocker overlays + panel renderer
    model_tasks.go              # Task list helpers (filter, sort, reload, export)
    model_features.go           # Feature handlers (focus, tags, undo, blockers)
    keys.go                     # KeyMap definitions (key.Binding)
    styles.go                   # Lip Gloss styles (cyan accent dark theme)
    delegate.go                 # Custom list.ItemDelegate for task rendering
    delegate_journal.go         # Custom list.ItemDelegate for journal notes
  cli/
    cli.go                      # Cobra root command + global flags
    output.go                   # Styled output (Printer, TTY-aware tables, colors)
    errors.go                   # NotFoundError type with errors.As support
    confirm.go                  # Confirmation prompts (styled when TTY)
    tasks.go                    # add, done, list, show, edit, delete, status
    batch.go                    # batch (stdin newline-delimited JSON commands)
    note.go                     # note (add, list, edit, delete)
    journal.go                  # journal (add, list, show, edit, delete, hide)
    export.go                   # export (md, json, file output with buffered flush)
    subtasks.go                 # subtask (add, list, done, edit, delete)
    timelog.go                  # timelog (add, list, summary)
    recur.go                    # recur (set, clear)
    focus.go                    # focus (start, status, stats)
    stats.go                    # stats (task + focus summary)
    config_cmd.go               # config (list, get, set, reset)
    completion.go               # Shell completion (bash, zsh, fish, powershell)
  config/
    config.go                   # JSON config (~/.todo-app/config.json)
  database/
    db.go                       # SQLite connection (WAL mode) + daily backup
  export/
    export.go                   # Markdown + JSON export writers
  focus/
    focus.go                    # Focus/Pomodoro session model
    store.go                    # Focus session SQLite repository
  journal/
    journal.go                  # Domain model (Note, Entry, DateTitle)
    store.go                    # SQLite repository (CRUD, batch queries, transactions)
  task/
    task.go                     # Domain model (Task, Subtask, Status, Priority)
    store.go                    # SQLite repository (CRUD, subtasks, tags, time logs)
    recur.go                    # Recurring task logic + next due date calculation
    timelog.go                  # Time log model + duration parsing/formatting
  ui/
    colors.go                   # Shared color palette (adaptive light/dark)
    views.go                    # View rendering (tabs, detail, status bar, dialogs)
    form.go                     # Huh form builders + form data types
```

### Data Model
```go
// Task management
type Task struct {
    ID, Title, Description, Status, Priority
    DueDate, CreatedAt, UpdatedAt
    RecurFreq, RecurInterval        // Recurrence (daily, weekly, monthly, yearly)
    BlockedByIDs []int64            // Tasks that block this one
    BlocksIDs    []int64            // Tasks this one blocks
    Subtasks     []Subtask
    Tags         []string
    Metadata     map[string]string  // Structured key-value pairs (JSON column)
    TimeLogs     []TimeLog
    Notes        []TaskNote         // Timestamped comments (task_notes table)
}

type TaskNote struct {
    ID, TaskID, Body, CreatedAt
}

type Subtask struct {
    ID, Title, Completed, Position
}

type TimeLog struct {
    ID, TaskID, Duration, Note, LoggedAt
}

// Journal
type Note struct {
    ID, Date, Hidden, CreatedAt, UpdatedAt
    Entries []Entry
}

type Entry struct {
    ID, NoteID, Body, CreatedAt
}

// Focus (Pomodoro)
type Session struct {
    ID, TaskID, Duration, StartedAt, CompletedAt
    Kind (Work/ShortBreak/LongBreak), CyclePos
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
- **Batch queries**: `List()` batch-loads all 6 relations (subtasks, tags, time logs, notes, blockerIDs, blocksIDs) using `WHERE IN` queries (7 total vs 6N+1); journal entries loaded similarly
- **Transactions**: All multi-statement writes use `db.Begin()`/`tx.Commit()`
- **Date parsing**: `time.ParseInLocation` for date-only fields (timezone-correct)
- **Printer pattern**: `cli/output.go` — `Printer` struct with `noColor`/`quiet` flags; methods: `Table()`, `Success()`, `Bold()`, `Dim()`, `Colored()`, `JSON()`
- **TTY detection**: `isTTY()` via `go-isatty` (not `ModeCharDevice`); auto-disables color when piped unless `--no-color` explicitly set
- **Styled tables**: lipgloss/table with `RoundedBorder()` when TTY, tabwriter fallback when piped; `writerWidth()` constrains to terminal width
- **Batch mode**: `rondo batch` reads newline-delimited JSON from stdin; creates a fresh cobra command tree per invocation to prevent flag state leaking; sub-command stdout/stderr suppressed (only JSON results emitted); rejects nested `batch` commands
- **UTC timestamps**: Storage timestamps (`created_at`, `updated_at`, focus sessions) use `.UTC()`; user-facing date operations (journal `DateTitle()`, `GetOrCreateToday()`, `parseJournalDate()`) use local time
- **Metadata storage**: JSON `TEXT` column with `marshalMetadata()`/`parseMetadata()` helpers; `addColumnIfNotExists()` migration
- **Blocker enrichment**: JSON output has `blocked_by`/`blocks` as `[]int64` (backward compat) plus `blocked_by_detail`/`blocks_detail` as `[]{id, title, status}` objects; resolved via `taskIndex` lookup
- **Delete guard**: CLI checks `ListBlocksIDs()` before delete; exit code 1 + error listing blocked tasks unless `--cascade` is passed; TUI uses `deleteGuardConfirmed` flag for two-step confirmation with yellow border
- **Transactional blockers**: `SetBlocksIDs()` replaces all tasks a blocker blocks in a single transaction; used by both TUI and CLI edit paths

### SQLite Schema
Database at `~/.todo-app/todo.db` with tables:
- `tasks`, `subtasks`, `tags`, `task_blocks`, `task_notes` — task management (ON DELETE CASCADE)
- `time_logs` — time tracking per task
- `journal_notes`, `journal_entries` — journal (ON DELETE CASCADE, indexed)
- `focus_sessions` — Pomodoro session tracking

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
- **Fix All Failures**: If a test or problem is found, fix it regardless of when it was introduced — no ignoring pre-existing issues
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

## Ecosystem Links
- Bubbletea: https://github.com/charmbracelet/bubbletea
- Bubbles: https://github.com/charmbracelet/bubbles
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Huh: https://github.com/charmbracelet/huh
- Cobra: https://github.com/spf13/cobra
