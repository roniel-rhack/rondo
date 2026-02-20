package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/roniel/todo-app/internal/export"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
)

// cmdExport handles `rondo export [flags]`.
func cmdExport(args []string, taskStore *task.Store, journalStore *journal.Store) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	format := fs.String("format", "md", "Export format: md, json")
	output := fs.String("output", "", "Output file path (default: stdout)")
	includeJournal := fs.Bool("journal", false, "Include journal entries")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Load tasks.
	tasks, err := taskStore.List()
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	// Optionally load journal notes.
	var notes []journal.Note
	if *includeJournal {
		notes, err = journalStore.ListNotes(false)
		if err != nil {
			return fmt.Errorf("list journal notes: %w", err)
		}
	}

	// Determine output destination.
	w := os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	switch *format {
	case "md", "markdown":
		if err := export.WriteTasks(w, tasks); err != nil {
			return fmt.Errorf("write tasks: %w", err)
		}
		if *includeJournal {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			if err := export.WriteNotes(w, notes); err != nil {
				return fmt.Errorf("write journal: %w", err)
			}
		}
	case "json":
		if !*includeJournal {
			notes = nil
		}
		if err := export.WriteJSON(w, tasks, notes); err != nil {
			return fmt.Errorf("write json: %w", err)
		}
	default:
		return fmt.Errorf("invalid format %q: must be md or json", *format)
	}

	if *output != "" {
		fmt.Fprintf(os.Stderr, "Exported to %s\n", *output)
	}
	return nil
}
