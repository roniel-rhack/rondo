package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/app"
	"github.com/roniel/todo-app/internal/cli"
	"github.com/roniel/todo-app/internal/config"
	"github.com/roniel/todo-app/internal/database"
	"github.com/roniel/todo-app/internal/focus"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
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

	// Load config early so it's available to both CLI and TUI.
	cfg, warnings, err := config.LoadWithWarnings()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load failed: %v\n", err)
		cfg = config.DefaultConfig()
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}

	// CLI subcommands: if args are provided, dispatch to CLI instead of TUI.
	if len(os.Args) > 1 {
		if err := cli.Run(os.Args[1:], taskStore, journalStore, focusStore, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if cli.IsNotFound(err) {
				os.Exit(3)
			}
			os.Exit(1)
		}
		return
	}

	// Detect terminal background and initialize color theme.
	ui.InitTheme(lipgloss.HasDarkBackground())

	m := app.New(taskStore, journalStore, focusStore, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
