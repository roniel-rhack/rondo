<h1 align="center">
  RonDO
</h1>

<p align="center">
  <strong>A modern terminal productivity app that combines task management with a daily journal.</strong>
</p>

<p align="center">
  <a href="https://github.com/roniel-rhack/rondo/releases/latest"><img src="https://img.shields.io/github/v/tag/roniel-rhack/rondo?style=flat-square&amp;label=release&amp;color=00bcd4" alt="Release"></a>
  <a href="https://github.com/roniel-rhack/rondo/actions"><img src="https://img.shields.io/github/actions/workflow/status/roniel-rhack/rondo/release.yml?style=flat-square&amp;label=CI" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/roniel-rhack/rondo"><img src="https://goreportcard.com/badge/github.com/roniel-rhack/rondo?style=flat-square" alt="Go Report Card"></a>
  <a href="https://github.com/roniel-rhack/rondo/blob/main/LICENSE"><img src="https://img.shields.io/github/license/roniel-rhack/rondo?style=flat-square&amp;color=00bcd4" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&amp;logo=go&amp;logoColor=white" alt="Go"></a>
</p>

<p align="center">
  RonDO is a single-binary terminal app for developers who want distraction-free productivity<br>
  without leaving their terminal. Tasks, journal, and Pomodoro timer in one keyboard-driven<br>
  interface — backed by local SQLite. No accounts, no cloud, no config required.
</p>

---

<p align="center">
  <img src="assets/tasks.png" width="720" alt="Task management view">
</p>
<p align="center"><em>Task management with subtasks, priorities, and time tracking</em></p>

<p align="center">
  <img src="assets/journal.png" width="720" alt="Journal view">
</p>
<p align="center"><em>Daily journal with timestamped entries and smart date labels</em></p>

---

## Contents

- [Install](#install)
- [Quick Start](#quick-start)
- [Features](#features)
- [CLI](#cli-mode)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Data & Config](#data--config)
- [Development](#development)
- [Architecture](#architecture)
- [License](#license)

## Install

### Go

```bash
go install github.com/roniel-rhack/rondo/cmd/todo@latest
```

### Homebrew

```bash
brew tap roniel-rhack/tap
brew install rondo
```

### Arch Linux (AUR)

```bash
yay -S rondo
```

### From source

```bash
git clone https://github.com/roniel-rhack/rondo.git
cd rondo
go build -o rondo ./cmd/todo
mv rondo /usr/local/bin/
```

## Quick Start

```bash
rondo                              # Launch the TUI
rondo add "My first task"          # Add a task from the CLI
rondo journal "Getting started"    # Write a journal entry
rondo list                         # See all tasks
```

All data is stored locally at `~/.todo-app/` — no setup needed.

## Features

### Task Management

- **Full CRUD** — create, view, edit, and delete tasks with validated forms
- **Subtasks** — completion tracking with progress bar
- **Status workflow** — Pending, In Progress, Done
- **Priority levels** — Low, Medium, High, Urgent (color-coded)
- **Due dates** — with overdue detection and sort support
- **Tags** — comma-separated, filterable
- **Recurring tasks** — daily, weekly, monthly, or yearly; auto-spawns next on completion
- **Task dependencies** — mark tasks as blocked by others
- **Time logging** — log time spent with optional notes
- **Sorting** — by creation date, due date, or priority
- **Fuzzy search** — instant filter across tasks

### Productivity Tools

- **Pomodoro timer** — full work/break cycles with configurable durations
  - Work → Short Break → Work → ... → Long Break (4-session sets)
  - Phase-aware: work, short break, long break with distinct indicators
  - Cycle progress indicator, terminal bell on completion
  - Configurable via settings form or `config.json`
- **Statistics overlay** — task counts, priority breakdown, focus sessions, streaks
- **Export** — Markdown or JSON, with optional journal inclusion
- **Undo** — revert the last destructive action

### Daily Journal

- **One note per day** — auto-created on first entry
- **Timestamped entries** — each entry records the time
- **Edit & delete entries** — cursor-based selection
- **Hide/restore notes** — archive without deleting
- **Smart date labels** — "Today", "Yesterday", weekday names

### Interface

- **Two-panel layout** — list + detail with resizable split
- **Four tabs** — All, Active, Done, Journal (live counts)
- **Vim-style navigation** — `j`/`k` everywhere
- **Context-sensitive status bar** — keybinding hints update per panel
- **Modal forms** — validated input with Dracula theme
- **Confirmation dialogs** — for all destructive actions
- **Help overlay** — press `?` for the full keybinding reference
- **Auto backups** — daily SQLite backups

## CLI Mode

Full-featured CLI with styled terminal output (auto-detected), JSON support, and shell completions.

#### Global Flags

| Flag | Description |
|------|-------------|
| `--format table\|json\|plain` | Output format (default: table) |
| `--json` | Shorthand for `--format json` |
| `-q, --quiet` | Suppress non-essential output |
| `--no-color` | Disable ANSI colors (auto-detected when piped) |

#### Basic Usage

```bash
rondo add "Buy groceries" --priority high --due 2026-03-15
rondo list --status pending --sort priority
rondo done 3
rondo show 3
rondo journal "Productive day"
rondo stats
```

<details>
<summary><strong>All CLI commands</strong></summary>

```bash
# Tasks
rondo add "Buy groceries" --priority high --due 2026-03-15 --tags "home,shopping"
rondo list --status pending --sort priority --limit 10
rondo list --priority urgent --overdue --format json
rondo show 3
rondo edit 3 --title "Buy organic groceries" --due 2026-03-20
rondo done 3 4 5
rondo delete 3 --force
rondo status 3 active

# Subtasks
rondo subtask add 3 "Pick up milk"
rondo subtask list 3
rondo subtask done 3 1

# Time tracking
rondo timelog add 3 1h30m --note "Deep work session"
rondo timelog list 3
rondo timelog summary --days 30

# Recurrence
rondo recur set 3 weekly
rondo recur clear 3

# Journal
rondo journal "Productive day"
rondo journal add "Wrapped up the feature" --date yesterday
rondo journal list
rondo journal show today

# Focus (Pomodoro)
rondo focus start --task-id 3 --duration 25m
rondo focus status
rondo focus stats --days 14

# Utilities
rondo stats
rondo export --format json --journal --output backup.json
rondo config list
rondo config set focus.work_duration_min 30
rondo completion zsh
```

</details>

## Keyboard Shortcuts

> **Start here:** `a` to add, `j`/`k` to navigate, `s` to change status, `Tab` to switch tabs, `?` for help.

### Global

| Key | Action |
|-----|--------|
| `Tab` | Switch tabs |
| `1` / `2` | Focus left / right panel |
| `<` / `>` | Resize panels |
| `Esc` | Return to list / clear filter |
| `?` | Help overlay |
| `p` | Pomodoro timer |
| `P` | Pomodoro settings |
| `G` | Statistics |
| `X` | Export |
| `Ctrl+Z` | Undo last action |
| `q` | Quit |

<details>
<summary><strong>Panel-specific shortcuts</strong></summary>

### Tasks — Panel 1

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate |
| `a` | Add task |
| `e` | Edit task |
| `d` | Delete task |
| `s` | Cycle status |
| `t` | Add subtask |
| `/` | Search |
| `F1`/`F2`/`F3` | Sort by created / due / priority |
| `F4` | Toggle tag filter bar |

### Tasks — Panel 2 (Details)

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate subtasks |
| `a` | Add subtask |
| `e` | Edit subtask |
| `d` | Delete subtask |
| `s` | Toggle subtask |
| `l` | Log time |
| `b` | View blockers |

### Journal — Panel 1 (Notes)

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate notes |
| `a` | Add entry (today) |
| `h` | Hide / restore note |
| `H` | Toggle show hidden |
| `/` | Search notes |

### Journal — Panel 2 (Entries)

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate entries |
| `a` | Add entry (today) |
| `e` | Edit entry |
| `d` | Delete entry |

</details>

## Data & Config

| Path | Purpose |
|------|---------|
| `~/.todo-app/todo.db` | SQLite database (WAL mode) |
| `~/.todo-app/config.json` | Persistent settings |
| `~/.todo-app/backups/` | Daily auto-backups |

## Development

Requires **Go 1.23+**.

```bash
go build -o rondo ./cmd/todo   # Build
go run ./cmd/todo              # Run
go test ./...                  # Test
go vet ./...                   # Vet
go mod tidy                    # Tidy deps
```

Contributions welcome — please open an issue first for feature discussions.

<details>
<summary><strong>Architecture</strong></summary>

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
    keys.go                     # Keybinding definitions
    styles.go                   # Lip Gloss styles
    delegate.go                 # Task list item delegate
    delegate_journal.go         # Journal note list item delegate
  cli/
    cli.go                      # Cobra root command + global flags
    output.go                   # Styled output (TTY-aware tables, colors)
    errors.go                   # NotFoundError type
    confirm.go                  # Confirmation prompts
    tasks.go                    # add, done, list, show, edit, delete, status
    journal.go                  # journal (add, list, show, edit, delete, hide)
    export.go                   # export (md, json, file output)
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
    db.go                       # SQLite connection + daily backup
    backup.go                   # Backup rotation logic
  export/
    export.go                   # Markdown + JSON export writers
  focus/
    focus.go                    # Focus/Pomodoro session model
    store.go                    # Focus session SQLite repository
  journal/
    journal.go                  # Note & Entry domain types
    store.go                    # Journal SQLite repository
  task/
    task.go                     # Task & Subtask domain types
    store.go                    # Task SQLite repository
    deps.go                     # Task dependency cycle detection
    recur.go                    # Recurring task logic
    timelog.go                  # Time log model
  ui/
    colors.go                   # Shared color palette (adaptive light/dark)
    views.go                    # Rendering (tabs, detail, status bar, dialogs)
    form.go                     # Huh form builders
    markdown.go                 # Markdown rendering
    overdue.go                  # Due date classification
    stats.go                    # Sparkline rendering
```

Follows the **Bubbletea MVU** (Model-Update-View) pattern. All data persists in a single SQLite database at `~/.todo-app/todo.db` (WAL mode, single connection, `ON DELETE CASCADE`).

</details>

## Built With

[Bubbletea](https://github.com/charmbracelet/bubbletea) (TUI) ·
[Bubbles](https://github.com/charmbracelet/bubbles) (components) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) (styling) ·
[Huh](https://github.com/charmbracelet/huh) (forms) ·
[Cobra](https://github.com/spf13/cobra) (CLI) ·
[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (database)

## License

[MIT](LICENSE)
