# RonDO

A modern terminal productivity app that combines **task management** with a **daily journal** — all in a fast, keyboard-driven TUI.

Built with Go and the [Charm](https://charm.sh) ecosystem.

```
┌──────────────────────────────────────────────────────────────────┐
│  RonDO  │  All (7)  │  Active (4)  │  Done (3)  │  Journal (5) │
├────────────────────────┬─────────────────────────────────────────┤
│  1: Tasks              │  2: Details                             │
│  ▸ Fix login bug    H  │  Title        Fix login bug             │
│    Update docs      M  │  Status       ● In Progress             │
│    Add dark mode    L  │  Priority     High                      │
│    Deploy v2.0      U  │  Due          Feb 25, 2026              │
│    Write tests      M  │  Created      Feb 18, 2026              │
│                        │  Tags         backend, auth              │
│                        │                                         │
│                        │  Subtasks     1/3                        │
│                        │  ██████████░░░░░░░░░░░░░░               │
│                        │                                         │
│                        │  ▸ [x] Reproduce the issue              │
│                        │    [ ] Write the fix                    │
│                        │    [ ] Add regression test              │
├────────────────────────┴─────────────────────────────────────────┤
│  7 tasks | 4 done | 3 active   [1:Tasks] a:add e:edit d:del ?:help│
└──────────────────────────────────────────────────────────────────┘
```

## Features

### Task Management
- **Full CRUD** — Create, view, edit, and delete tasks with validated forms
- **Subtasks** — Add subtasks to any task, toggle completion, track progress
- **Status workflow** — Cycle between Pending → In Progress → Done
- **Priority levels** — Low, Medium, High, Urgent with color coding
- **Due dates** — Optional due date tracking with sort support
- **Tags** — Comma-separated tags for organization
- **Sorting** — Sort by creation date, due date, or priority (F1/F2/F3)
- **Fuzzy search** — Filter tasks by typing with `/`

### Daily Journal
- **One note per day** — Automatically creates today's note
- **Timestamped entries** — Each entry records the time it was written
- **Edit & delete entries** — Cursor-based selection for precise editing
- **Hide/restore notes** — Archive old notes without deleting them
- **Smart date labels** — "Today", "Yesterday", weekday names, or full dates

### Interface
- **Two-panel layout** — List on the left, details on the right
- **Four tabs** — All, Active, Done, Journal — with live counts
- **Vim-style navigation** — `j`/`k` everywhere, `1`/`2` for panel focus
- **Context-sensitive keys** — Status bar hints change based on focused panel
- **Modal forms** — Huh-powered forms with validation and Dracula theme
- **Confirmation dialogs** — All destructive actions require confirmation
- **Help overlay** — Press `?` for a complete keybinding reference

### Technical
- **Zero CGO** — Pure Go SQLite via `modernc.org/sqlite`
- **Single database** — All data in `~/.todo-app/todo.db`
- **WAL mode** — SQLite write-ahead logging for performance
- **Transactions** — All multi-statement writes are atomic
- **Batch queries** — Journal entries loaded without N+1 queries
- **Timezone-correct** — Date comparisons use local time, not UTC

## Installation

### From source

Requires Go 1.23 or later.

```bash
git clone https://github.com/roniel/todo-app.git
cd todo-app
go build -o rondo ./cmd/todo
```

Move the binary to your PATH:

```bash
mv rondo /usr/local/bin/
```

### Run directly

```bash
go run ./cmd/todo
```

## Usage

Launch the app:

```bash
rondo
```

### Quick Start

1. Press `a` to add your first task
2. Fill in the form fields, press `Enter` to submit
3. Press `s` to cycle the task status
4. Press `Tab` to switch between All / Active / Done / Journal tabs
5. Press `?` at any time for the full keybinding reference

### Keyboard Shortcuts

#### Global
| Key | Action |
|-----|--------|
| `Tab` | Switch tabs |
| `1` / `2` | Focus left / right panel |
| `Esc` | Return to list / clear filter |
| `?` | Help overlay |
| `q` | Quit |

#### Tasks (Panel 1)
| Key | Action |
|-----|--------|
| `j` / `k` | Navigate |
| `a` | Add task |
| `e` | Edit task |
| `d` | Delete task |
| `s` | Cycle status |
| `t` | Add subtask |
| `/` | Search |
| `F1`/`F2`/`F3` | Sort created / due / priority |

#### Task Details (Panel 2)
| Key | Action |
|-----|--------|
| `j` / `k` | Navigate subtasks |
| `a` | Add subtask |
| `e` | Edit subtask |
| `d` | Delete subtask |
| `s` | Toggle subtask |

#### Journal — Notes (Panel 1)
| Key | Action |
|-----|--------|
| `j` / `k` | Navigate notes |
| `a` | Add entry (today) |
| `h` | Hide / restore note |
| `H` | Toggle show hidden |
| `/` | Search notes |

#### Journal — Entries (Panel 2)
| Key | Action |
|-----|--------|
| `j` / `k` | Navigate entries |
| `e` | Edit entry |
| `d` | Delete entry |
| `a` | Add entry (today) |

## Architecture

```
cmd/todo/main.go                # Entry point
internal/
  app/
    model.go                    # Main Bubbletea Model + Update + View
    model_journal.go            # Journal tab handlers
    keys.go                     # Keybinding definitions
    styles.go                   # Lip Gloss styles
    delegate.go                 # Task list item delegate
    delegate_journal.go         # Journal note list item delegate
  database/
    db.go                       # SQLite connection setup
  journal/
    journal.go                  # Note & Entry domain types
    store.go                    # Journal SQLite repository
  task/
    task.go                     # Task & Subtask domain types
    store.go                    # Task SQLite repository
  ui/
    colors.go                   # Shared color palette
    views.go                    # Rendering (tabs, detail, status bar, dialogs)
    form.go                     # Huh form builders
```

The app follows the **Bubbletea MVU** (Model-Update-View) pattern:

- **Model** holds all application state: tasks, notes, lists, viewport, forms, mode, focus
- **Update** dispatches messages: global keys first, then per-tab handlers
- **View** renders the full UI: tab bar, split panels, status bar, modal overlays

### Data Storage

All data persists in a single SQLite database at `~/.todo-app/todo.db` with five tables:

| Table | Purpose |
|-------|---------|
| `tasks` | Task records (title, description, status, priority, dates) |
| `subtasks` | Subtask records linked to tasks |
| `tags` | Tag records linked to tasks |
| `journal_notes` | One row per calendar day |
| `journal_entries` | Timestamped entries linked to notes |

Foreign keys use `ON DELETE CASCADE`. The database runs in WAL mode with a single connection.

## Development

```bash
# Build
go build ./cmd/todo

# Run
go run ./cmd/todo

# Test
go test ./...

# Vet
go vet ./...

# Tidy dependencies
go mod tidy
```

## Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework (MVU)
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components (list, viewport, help)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions
- [Huh](https://github.com/charmbracelet/huh) — Terminal forms
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — CGO-free SQLite driver

## License

MIT
