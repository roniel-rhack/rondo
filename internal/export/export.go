package export

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
)

// WriteTasks writes tasks in Markdown format to w.
func WriteTasks(w io.Writer, tasks []task.Task) error {
	if _, err := fmt.Fprintln(w, "# Tasks"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if len(tasks) == 0 {
		_, err := fmt.Fprintln(w, "_No tasks._")
		return err
	}

	for _, t := range tasks {
		checkbox := "[ ]"
		if t.Status == task.Done {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("- %s **%s**", checkbox, t.Title)

		var meta []string
		meta = append(meta, t.Priority.String())
		meta = append(meta, t.Status.String())
		if t.DueDate != nil {
			meta = append(meta, "due "+t.DueDate.Format("2006-01-02"))
		}
		if len(t.Tags) > 0 {
			meta = append(meta, "tags: "+strings.Join(t.Tags, ", "))
		}
		if len(meta) > 0 {
			line += " (" + strings.Join(meta, " | ") + ")"
		}

		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}

		if t.Description != "" {
			if _, err := fmt.Fprintf(w, "  > %s\n", t.Description); err != nil {
				return err
			}
		}

		for _, st := range t.Subtasks {
			sub := "[ ]"
			if st.Completed {
				sub = "[x]"
			}
			if _, err := fmt.Fprintf(w, "  - %s %s\n", sub, st.Title); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteNotes writes journal notes in Markdown format to w.
func WriteNotes(w io.Writer, notes []journal.Note) error {
	if _, err := fmt.Fprintln(w, "# Journal"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if len(notes) == 0 {
		_, err := fmt.Fprintln(w, "_No journal entries._")
		return err
	}

	for i, n := range notes {
		if _, err := fmt.Fprintf(w, "## %s\n\n", n.Date.Format("2006-01-02")); err != nil {
			return err
		}
		if len(n.Entries) == 0 {
			if _, err := fmt.Fprintln(w, "_No entries._"); err != nil {
				return err
			}
		}
		for _, e := range n.Entries {
			ts := e.CreatedAt.Format("15:04")
			if _, err := fmt.Fprintf(w, "- **%s** %s\n", ts, e.Body); err != nil {
				return err
			}
		}
		if i < len(notes)-1 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
	}
	return nil
}

// exportData is the JSON structure for combined export.
type exportData struct {
	Tasks   []jsonTask   `json:"tasks"`
	Journal []jsonNote   `json:"journal,omitempty"`
}

type jsonTask struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	DueDate     string       `json:"due_date,omitempty"`
	CreatedAt   string       `json:"created_at"`
	Tags        []string     `json:"tags,omitempty"`
	Subtasks    []jsonSubtask `json:"subtasks,omitempty"`
}

type jsonSubtask struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type jsonNote struct {
	Date    string      `json:"date"`
	Entries []jsonEntry `json:"entries"`
}

type jsonEntry struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// WriteJSON writes tasks and optional journal notes as JSON to w.
func WriteJSON(w io.Writer, tasks []task.Task, notes []journal.Note) error {
	data := exportData{}

	for _, t := range tasks {
		jt := jsonTask{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Status:      t.Status.String(),
			Priority:    t.Priority.String(),
			CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tags:        t.Tags,
		}
		if t.DueDate != nil {
			jt.DueDate = t.DueDate.Format("2006-01-02")
		}
		for _, st := range t.Subtasks {
			jt.Subtasks = append(jt.Subtasks, jsonSubtask{
				ID:        st.ID,
				Title:     st.Title,
				Completed: st.Completed,
			})
		}
		data.Tasks = append(data.Tasks, jt)
	}

	if notes != nil {
		for _, n := range notes {
			jn := jsonNote{
				Date: n.Date.Format("2006-01-02"),
			}
			for _, e := range n.Entries {
				jn.Entries = append(jn.Entries, jsonEntry{
					Body:      e.Body,
					CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
			data.Journal = append(data.Journal, jn)
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
