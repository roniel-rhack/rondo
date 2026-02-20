package cli

import (
	"fmt"

	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
)

// Run dispatches a CLI subcommand based on the first element of args.
// Supported subcommands: add, done, list, journal, export.
func Run(args []string, taskStore *task.Store, journalStore *journal.Store) error {
	if len(args) == 0 {
		return fmt.Errorf("no subcommand provided. Available: add, done, list, journal, export")
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "add":
		return cmdAdd(rest, taskStore)
	case "done":
		return cmdDone(rest, taskStore)
	case "list":
		return cmdList(rest, taskStore)
	case "journal":
		return cmdJournalAdd(rest, journalStore)
	case "export":
		return cmdExport(rest, taskStore, journalStore)
	default:
		return fmt.Errorf("unknown subcommand %q. Available: add, done, list, journal, export", cmd)
	}
}
