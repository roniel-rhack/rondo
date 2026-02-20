package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/roniel/todo-app/internal/journal"
)

// cmdJournalAdd handles `rondo journal "entry text"`.
func cmdJournalAdd(args []string, store *journal.Store) error {
	fs := flag.NewFlagSet("journal", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: rondo journal \"entry text\"")
	}

	body := strings.Join(fs.Args(), " ")

	note, err := store.GetOrCreateToday()
	if err != nil {
		return fmt.Errorf("get today note: %w", err)
	}

	if err := store.AddEntry(note.ID, body); err != nil {
		return fmt.Errorf("add entry: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Added journal entry to %s\n", note.Date.Format("2006-01-02"))
	return nil
}
