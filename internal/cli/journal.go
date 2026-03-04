package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/roniel/todo-app/internal/journal"
	"github.com/spf13/cobra"
)

// parseJournalDate converts "today", "yesterday", or "YYYY-MM-DD" to a date
// string in time.DateOnly format.
func parseJournalDate(s string) (string, error) {
	switch strings.ToLower(s) {
	case "today", "":
		return time.Now().Format(time.DateOnly), nil
	case "yesterday":
		return time.Now().AddDate(0, 0, -1).Format(time.DateOnly), nil
	default:
		t, err := time.ParseInLocation(time.DateOnly, s, time.Local)
		if err != nil {
			return "", fmt.Errorf("invalid date %q: expected today, yesterday, or YYYY-MM-DD", s)
		}
		return t.Format(time.DateOnly), nil
	}
}

// findNoteByDate returns the note matching dateStr from the full note list.
func findNoteByDate(notes []journal.Note, dateStr string) (*journal.Note, bool) {
	for i := range notes {
		if notes[i].Date.Format(time.DateOnly) == dateStr {
			return &notes[i], true
		}
	}
	return nil, false
}

func (c *CLI) journalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "journal [\"entry text\"]",
		Short: "Manage journal entries",
		Long: `Manage journal entries.

When called with text arguments and no subcommand, adds an entry to today's
note (backward-compatible shorthand for 'journal add').`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no text provided; use 'rondo journal <text>' or a subcommand (run 'rondo journal --help')")
			}
			// Backward-compat: treat args as text to add to today.
			body := strings.Join(args, " ")
			note, err := c.journalStore.GetOrCreateToday()
			if err != nil {
				return fmt.Errorf("get today note: %w", err)
			}
			if err := c.journalStore.AddEntry(note.ID, body); err != nil {
				return fmt.Errorf("add entry: %w", err)
			}
			p := c.printer(os.Stdout)
			p.Success("Added journal entry to %s", c.cfg.FormatDate(note.Date))
			return nil
		},
	}

	cmd.AddCommand(c.journalAddCmd())
	cmd.AddCommand(c.journalListCmd())
	cmd.AddCommand(c.journalShowCmd())
	cmd.AddCommand(c.journalEditCmd())
	cmd.AddCommand(c.journalDeleteCmd())
	cmd.AddCommand(c.journalHideCmd())

	return cmd
}

func (c *CLI) journalAddCmd() *cobra.Command {
	var date string

	cmd := &cobra.Command{
		Use:   "add \"entry text\"",
		Short: "Add a journal entry",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := strings.Join(args, " ")

			var noteID int64
			var noteDate time.Time

			if date == "" {
				n, err := c.journalStore.GetOrCreateToday()
				if err != nil {
					return fmt.Errorf("get today note: %w", err)
				}
				noteID = n.ID
				noteDate = n.Date
			} else {
				dateStr, err := parseJournalDate(date)
				if err != nil {
					return err
				}
				n, err := c.journalStore.GetOrCreate(dateStr)
				if err != nil {
					return fmt.Errorf("get note for %s: %w", dateStr, err)
				}
				noteID = n.ID
				noteDate = n.Date
			}

			if err := c.journalStore.AddEntry(noteID, body); err != nil {
				return fmt.Errorf("add entry: %w", err)
			}

			p := c.printer(os.Stdout)
			p.Success("Added journal entry to %s", c.cfg.FormatDate(noteDate))
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Note date (today, yesterday, or YYYY-MM-DD; default: today)")

	return cmd
}

func (c *CLI) journalListCmd() *cobra.Command {
	var date string
	var hidden bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List journal notes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := c.journalStore.ListNotes(hidden)
			if err != nil {
				return fmt.Errorf("list notes: %w", err)
			}

			// Filter by date if provided.
			if date != "" {
				dateStr, err := parseJournalDate(date)
				if err != nil {
					return err
				}
				var filtered []journal.Note
				for _, n := range notes {
					if n.Date.Format(time.DateOnly) == dateStr {
						filtered = append(filtered, n)
					}
				}
				notes = filtered
			}

			switch strings.ToLower(c.format) {
			case "json":
				return printNotesJSON(os.Stdout, notes)
			default:
				return c.printNotesTable(c.printer(os.Stdout), notes)
			}
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Filter to a specific date (today, yesterday, or YYYY-MM-DD)")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "Include hidden notes")

	return cmd
}

func (c *CLI) journalShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [today|yesterday|YYYY-MM-DD]",
		Short: "Show entries for a journal note",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dateArg := "today"
			if len(args) > 0 {
				dateArg = args[0]
			}

			dateStr, err := parseJournalDate(dateArg)
			if err != nil {
				return err
			}

			notes, err := c.journalStore.ListNotes(true)
			if err != nil {
				return fmt.Errorf("list notes: %w", err)
			}

			n, found := findNoteByDate(notes, dateStr)
			if !found {
				return fmt.Errorf("no journal note found for %s", dateStr)
			}

			entries, err := c.journalStore.ListEntries(n.ID)
			if err != nil {
				return fmt.Errorf("list entries: %w", err)
			}

			switch strings.ToLower(c.format) {
			case "json":
				return printEntriesJSON(os.Stdout, n.Date, entries)
			default:
				return c.printEntriesTable(c.printer(os.Stdout), n.Date, entries)
			}
		},
	}
}

func (c *CLI) journalEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <entry-id> \"new text\"",
		Short: "Edit a journal entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || id <= 0 {
				return fmt.Errorf("invalid entry ID %q: expected a positive integer", args[0])
			}
			body := args[1]
			if err := c.journalStore.UpdateEntry(id, body); err != nil {
				return fmt.Errorf("update entry: %w", err)
			}
			p := c.printer(os.Stdout)
			p.Success("Updated entry #%d", id)
			return nil
		},
	}
}

func (c *CLI) journalDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <entry-id>",
		Short:   "Delete a journal entry",
		Aliases: []string{"del", "rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || id <= 0 {
				return fmt.Errorf("invalid entry ID %q: expected a positive integer", args[0])
			}

			ok, err := Confirm(fmt.Sprintf("Delete entry #%d?", id), force)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}

			if err := c.journalStore.DeleteEntry(id); err != nil {
				return fmt.Errorf("delete entry: %w", err)
			}

			p := c.printer(os.Stdout)
			p.Success("Deleted entry #%d", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "y", false, "Skip confirmation prompt")

	return cmd
}

func (c *CLI) journalHideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hide <date>",
		Short: "Toggle the hidden flag on a journal note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dateStr, err := parseJournalDate(args[0])
			if err != nil {
				return err
			}

			notes, err := c.journalStore.ListNotes(true)
			if err != nil {
				return fmt.Errorf("list notes: %w", err)
			}

			n, found := findNoteByDate(notes, dateStr)
			if !found {
				return fmt.Errorf("no journal note found for %s", dateStr)
			}

			if err := c.journalStore.ToggleHidden(n.ID); err != nil {
				return fmt.Errorf("toggle hidden: %w", err)
			}

			p := c.printer(os.Stdout)
			p.Success("Toggled hidden flag for note %s", dateStr)
			return nil
		},
	}
}

// --- Output helpers ---

type jsonNote struct {
	ID         int64  `json:"id"`
	Date       string `json:"date"`
	EntryCount int    `json:"entry_count"`
	Hidden     bool   `json:"hidden"`
}

type jsonEntry struct {
	ID        int64  `json:"id"`
	NoteID    int64  `json:"note_id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type jsonNoteShow struct {
	Date    string      `json:"date"`
	Entries []jsonEntry `json:"entries"`
}

func printNotesJSON(w io.Writer, notes []journal.Note) error {
	out := make([]jsonNote, 0, len(notes))
	for _, n := range notes {
		out = append(out, jsonNote{
			ID:         n.ID,
			Date:       n.Date.Format(time.DateOnly),
			EntryCount: len(n.Entries),
			Hidden:     n.Hidden,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func (c *CLI) printNotesTable(p *Printer, notes []journal.Note) error {
	now := time.Now()
	rows := make([][]string, 0, len(notes))
	for _, n := range notes {
		hidden := "no"
		if n.Hidden {
			hidden = "yes"
		}
		rows = append(rows, []string{c.cfg.FormatNoteTitle(n.Date, now), fmt.Sprintf("%d", len(n.Entries)), hidden})
	}
	p.Table([]string{"DATE", "ENTRIES", "HIDDEN"}, rows)
	return nil
}

func printEntriesJSON(w io.Writer, date time.Time, entries []journal.Entry) error {
	out := jsonNoteShow{Date: date.Format(time.DateOnly)}
	for _, e := range entries {
		out.Entries = append(out.Entries, jsonEntry{
			ID:        e.ID,
			NoteID:    e.NoteID,
			Body:      e.Body,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		})
	}
	if out.Entries == nil {
		out.Entries = []jsonEntry{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func (c *CLI) printEntriesTable(p *Printer, date time.Time, entries []journal.Entry) error {
	fmt.Fprintf(p.w, "%s %s\n\n", p.Bold("Date:"), c.cfg.FormatDate(date))
	if len(entries) == 0 {
		fmt.Fprintln(p.w, p.Dim("(no entries)"))
		return nil
	}
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			fmt.Sprintf("%d", e.ID),
			c.cfg.FormatTime(e.CreatedAt),
			e.Body,
		})
	}
	p.Table([]string{"ID", "TIME", "BODY"}, rows)
	return nil
}
