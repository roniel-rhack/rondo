package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/roniel/todo-app/internal/app"
	"github.com/roniel/todo-app/internal/database"
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

	m := app.New(taskStore, journalStore)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
