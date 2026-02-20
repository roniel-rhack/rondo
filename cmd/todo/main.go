package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/roniel/todo-app/internal/app"
	"github.com/roniel/todo-app/internal/cli"
	"github.com/roniel/todo-app/internal/config"
	"github.com/roniel/todo-app/internal/database"
	"github.com/roniel/todo-app/internal/focus"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
)

func main() {
	db, err := database.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Auto backup on startup (best-effort, don't block).
	home, _ := os.UserHomeDir()
	if home != "" {
		backupDir := filepath.Join(home, ".todo-app", "backups")
		if err := database.Backup(db, backupDir, 30); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: backup failed: %v\n", err)
		}
	}

	taskStore, err := task.NewStore(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	journalStore, err := journal.NewStore(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	focusStore, err := focus.NewStore(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// CLI subcommands: if args are provided, dispatch to CLI instead of TUI.
	if len(os.Args) > 1 {
		if err := cli.Run(os.Args[1:], taskStore, journalStore); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load config.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load failed: %v\n", err)
		cfg = config.DefaultConfig()
	}

	m := app.New(taskStore, journalStore, focusStore, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
