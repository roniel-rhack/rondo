package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/roniel/todo-app/internal/task"
	"github.com/spf13/cobra"
)

func (c *CLI) noteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "Manage task notes",
	}
	cmd.AddCommand(
		c.noteAddCmd(),
		c.noteListCmd(),
		c.noteEditCmd(),
		c.noteDeleteCmd(),
	)
	return cmd
}

func (c *CLI) noteAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <task-id> \"note text\"",
		Short: "Add a note to a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}
			if _, err := c.getTaskOrNotFound(taskID); err != nil {
				return err
			}
			body := strings.TrimSpace(args[1])
			if body == "" {
				return fmt.Errorf("note body cannot be empty")
			}
			if err := c.taskStore.AddNote(taskID, body); err != nil {
				return fmt.Errorf("add note: %w", err)
			}
			c.printer(os.Stdout).Success("Added note to task #%d", taskID)
			return nil
		},
	}
}

func (c *CLI) noteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <task-id>",
		Short: "List notes for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}
			if _, err := c.getTaskOrNotFound(taskID); err != nil {
				return err
			}
			notes, err := c.taskStore.ListNotes(taskID)
			if err != nil {
				return fmt.Errorf("list notes: %w", err)
			}
			p := c.printer(os.Stdout)
			if strings.ToLower(c.format) == "json" {
				type jsonN struct {
					ID        int64  `json:"id"`
					Body      string `json:"body"`
					CreatedAt string `json:"created_at"`
				}
				out := make([]jsonN, 0, len(notes))
				for _, n := range notes {
					out = append(out, jsonN{
						ID:        n.ID,
						Body:      n.Body,
						CreatedAt: n.CreatedAt.Format(time.RFC3339),
					})
				}
				return p.JSON(out)
			}
			rows := make([][]string, 0, len(notes))
			for _, n := range notes {
				rows = append(rows, []string{
					strconv.FormatInt(n.ID, 10),
					c.cfg.FormatDateTime(n.CreatedAt),
					n.Body,
				})
			}
			p.Table([]string{"ID", "DATE", "NOTE"}, rows)
			return nil
		},
	}
}

func (c *CLI) noteEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <task-id> <note-id> \"new text\"",
		Short: "Edit a note",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}
			noteID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note ID %q: %w", args[1], err)
			}
			newBody := args[2]
			t, err := c.getTaskOrNotFound(taskID)
			if err != nil {
				return err
			}
			if !noteBelongsToTask(t, noteID) {
				return &NotFoundError{Type: "note", ID: noteID}
			}
			if err := c.taskStore.UpdateNote(noteID, newBody); err != nil {
				return fmt.Errorf("update note: %w", err)
			}
			c.printer(os.Stdout).Success("Updated note #%d", noteID)
			return nil
		},
	}
}

func (c *CLI) noteDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <task-id> <note-id>",
		Short: "Delete a note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}
			noteID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note ID %q: %w", args[1], err)
			}
			t, err := c.getTaskOrNotFound(taskID)
			if err != nil {
				return err
			}
			var noteBody string
			found := false
			for _, n := range t.Notes {
				if n.ID == noteID {
					noteBody = n.Body
					found = true
					break
				}
			}
			if !found {
				return &NotFoundError{Type: "note", ID: noteID}
			}
			// Truncate body for confirmation prompt
			display := noteBody
			if len(display) > 50 {
				display = display[:50] + "..."
			}
			ok, err := Confirm(fmt.Sprintf("Delete note #%d %q?", noteID, display), force)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}
			if err := c.taskStore.DeleteNote(noteID); err != nil {
				return fmt.Errorf("delete note: %w", err)
			}
			c.printer(os.Stdout).Success("Deleted note #%d", noteID)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "y", false, "Skip confirmation prompt")
	return cmd
}

// noteBelongsToTask reports whether noteID is among the task's notes.
func noteBelongsToTask(t *task.Task, noteID int64) bool {
	for _, n := range t.Notes {
		if n.ID == noteID {
			return true
		}
	}
	return false
}
