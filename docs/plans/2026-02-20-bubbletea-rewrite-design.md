# Todo App - Go + Bubbletea Rewrite Design

## Goal

Rewrite the todo TUI app from Java/TamboUI to Go using the Charm ecosystem (Bubbletea + Bubbles + Lip Gloss + Huh) with SQLite persistence. Replace the existing Java project entirely.

## Architecture

**Pattern:** Bubbletea MVU (Model-Update-View)

**Stack:**
- Bubbletea v1.3.x - main loop
- Bubbles v1.0.x - list, help, viewport, progress, key
- Lip Gloss v1.1.x - styling
- Huh v0.8.x - forms
- modernc.org/sqlite - CGO-free SQLite

**File structure:**
```
cmd/todo/main.go           # Entry point
internal/
  app/
    model.go               # Main model + Update/View
    keys.go                # KeyMap definitions
    styles.go              # Lip Gloss styles (theme)
    delegate.go            # Custom list.ItemDelegate
  task/
    task.go                # Domain model
    store.go               # SQLite repository
  ui/
    views.go               # View rendering functions
    form.go                # Huh form builders
go.mod / go.sum
```

## UI Layout

Two-panel layout with header tabs and footer help bar.

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

Task forms use Huh with fields: Title, Description, Priority (select), Due Date, Tags.

## Data Model

SQLite database at `~/.todo-app/todo.db`.

**Tables:** tasks, subtasks, tags (with ON DELETE CASCADE).

**Domain types:** Task (id, title, description, status, priority, due_date, created_at, updated_at, subtasks, tags), Subtask (id, title, completed, position), Status enum (Pending/InProgress/Done), Priority enum (Low/Medium/High/Urgent).

## Keybindings

| Key | Action |
|-----|--------|
| j/k/arrows | Navigate |
| Enter | Select |
| a | Add task (Huh form) |
| e | Edit task (Huh form) |
| d | Delete (confirm) |
| s | Cycle status |
| t | Add subtask |
| x | Toggle subtask |
| / | Search (fuzzy) |
| Tab | Switch tab |
| F1-F3 | Sort |
| ? | Toggle help |
| q/ctrl+c | Quit |

## Decisions

- Use Bubbletea v1 (stable) not v2 (beta)
- SQLite via modernc.org/sqlite (no CGO, single binary)
- Huh for forms (better UX than manual bubbles/textinput)
- Replace Java project in-place (delete Java files, init Go module)
- Data file: `~/.todo-app/todo.db`
