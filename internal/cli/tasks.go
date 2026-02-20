package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/roniel/todo-app/internal/task"
)

// cmdAdd handles `rondo add "task title" [flags]`.
func cmdAdd(args []string, store *task.Store) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	priority := fs.String("priority", "low", "Priority: low, medium, high, urgent")
	due := fs.String("due", "", "Due date (YYYY-MM-DD)")
	tags := fs.String("tags", "", "Comma-separated tags")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: rondo add \"task title\" [--priority low|medium|high|urgent] [--due YYYY-MM-DD] [--tags tag1,tag2]")
	}

	title := fs.Arg(0)

	t := &task.Task{
		Title: title,
	}

	// Parse priority.
	switch strings.ToLower(*priority) {
	case "low":
		t.Priority = task.Low
	case "medium", "med":
		t.Priority = task.Medium
	case "high":
		t.Priority = task.High
	case "urgent":
		t.Priority = task.Urgent
	default:
		return fmt.Errorf("invalid priority %q: must be low, medium, high, or urgent", *priority)
	}

	// Parse due date.
	if *due != "" {
		d, err := time.ParseInLocation(time.DateOnly, *due, time.Local)
		if err != nil {
			return fmt.Errorf("invalid due date %q: expected YYYY-MM-DD", *due)
		}
		t.DueDate = &d
	}

	// Parse tags.
	if *tags != "" {
		for _, tag := range strings.Split(*tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				t.Tags = append(t.Tags, tag)
			}
		}
	}

	if err := store.Create(t); err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Created task #%d: %s\n", t.ID, t.Title)
	return nil
}

// cmdDone handles `rondo done <task-id>`.
func cmdDone(args []string, store *task.Store) error {
	fs := flag.NewFlagSet("done", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: rondo done <task-id>")
	}

	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID %q: %w", fs.Arg(0), err)
	}

	// List all tasks to find the one we need.
	tasks, err := store.List()
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	var found *task.Task
	for i := range tasks {
		if tasks[i].ID == id {
			found = &tasks[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("task #%d not found", id)
	}

	found.Status = task.Done
	if err := store.Update(found); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Marked task #%d as done: %s\n", found.ID, found.Title)
	return nil
}

// cmdList handles `rondo list [flags]`.
func cmdList(args []string, store *task.Store) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	status := fs.String("status", "all", "Filter: pending, active, done, all")
	format := fs.String("format", "table", "Output format: table, json")

	if err := fs.Parse(args); err != nil {
		return err
	}

	tasks, err := store.List()
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	// Filter by status.
	filtered := filterTasks(tasks, *status)

	switch strings.ToLower(*format) {
	case "json":
		return printTasksJSON(filtered)
	case "table":
		return printTasksTable(filtered)
	default:
		return fmt.Errorf("invalid format %q: must be table or json", *format)
	}
}

func filterTasks(tasks []task.Task, status string) []task.Task {
	switch strings.ToLower(status) {
	case "pending":
		return filterByStatus(tasks, task.Pending)
	case "active", "in-progress", "inprogress":
		return filterByStatus(tasks, task.InProgress)
	case "done":
		return filterByStatus(tasks, task.Done)
	default:
		return tasks
	}
}

func filterByStatus(tasks []task.Task, s task.Status) []task.Task {
	var out []task.Task
	for _, t := range tasks {
		if t.Status == s {
			out = append(out, t)
		}
	}
	return out
}

func printTasksTable(tasks []task.Task) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tPRIORITY\tTITLE\tDUE\tTAGS")
	fmt.Fprintln(w, "--\t------\t--------\t-----\t---\t----")

	for _, t := range tasks {
		due := "-"
		if t.DueDate != nil {
			due = t.DueDate.Format("2006-01-02")
		}
		tags := "-"
		if len(t.Tags) > 0 {
			tags = strings.Join(t.Tags, ", ")
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			t.ID, t.Status.String(), t.Priority.String(), t.Title, due, tags)
	}
	return w.Flush()
}

func printTasksJSON(tasks []task.Task) error {
	type jsonTask struct {
		ID       int64    `json:"id"`
		Title    string   `json:"title"`
		Status   string   `json:"status"`
		Priority string   `json:"priority"`
		DueDate  string   `json:"due_date,omitempty"`
		Tags     []string `json:"tags,omitempty"`
	}

	out := make([]jsonTask, 0, len(tasks))
	for _, t := range tasks {
		jt := jsonTask{
			ID:       t.ID,
			Title:    t.Title,
			Status:   t.Status.String(),
			Priority: t.Priority.String(),
			Tags:     t.Tags,
		}
		if t.DueDate != nil {
			jt.DueDate = t.DueDate.Format("2006-01-02")
		}
		out = append(out, jt)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
