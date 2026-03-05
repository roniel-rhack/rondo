package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
	"github.com/spf13/cobra"
)

// parsePriority converts a string priority to a task.Priority value.
func parsePriority(s string) (task.Priority, error) {
	switch strings.ToLower(s) {
	case "low":
		return task.Low, nil
	case "medium", "med":
		return task.Medium, nil
	case "high":
		return task.High, nil
	case "urgent":
		return task.Urgent, nil
	default:
		return task.Low, fmt.Errorf("invalid priority %q: must be low, medium, high, or urgent", s)
	}
}

// getTaskOrNotFound retrieves a task by ID, wrapping sql.ErrNoRows as NotFoundError.
func (c *CLI) getTaskOrNotFound(id int64) (*task.Task, error) {
	t, err := c.taskStore.GetByID(id)
	if err == sql.ErrNoRows {
		return nil, &NotFoundError{Type: "task", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("get task %d: %w", id, err)
	}
	return t, nil
}

// parseMetaFlags parses --meta key=value pairs into a map.
func parseMetaFlags(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(raw))
	for _, kv := range raw {
		eq := strings.IndexByte(kv, '=')
		if eq < 1 {
			return nil, fmt.Errorf("invalid --meta %q: expected key=value", kv)
		}
		m[kv[:eq]] = kv[eq+1:]
	}
	return m, nil
}

// parseBlocksFlag parses a comma-separated list of task IDs.
func parseBlocksFlag(s string) ([]int64, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid task ID %q in --blocks: %w", p, err)
		}
		if id <= 0 {
			return nil, fmt.Errorf("invalid task ID %q in --blocks: must be positive", p)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *CLI) addCmd() *cobra.Command {
	var priority, due, tags, desc, recur, blocks string
	var meta []string

	cmd := &cobra.Command{
		Use:   "add \"task title\" [flags]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			t := &task.Task{Title: title, Description: desc}

			prio, err := parsePriority(priority)
			if err != nil {
				return err
			}
			t.Priority = prio

			if due != "" {
				d, err := time.ParseInLocation(time.DateOnly, due, time.UTC)
				if err != nil {
					return fmt.Errorf("invalid due date %q: expected YYYY-MM-DD", due)
				}
				t.DueDate = &d
			}

			if tags != "" {
				for _, tag := range strings.Split(tags, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						t.Tags = append(t.Tags, tag)
					}
				}
			}

			if m, err := parseMetaFlags(meta); err != nil {
				return err
			} else {
				t.Metadata = m
			}

			if err := c.taskStore.Create(t); err != nil {
				return fmt.Errorf("create task: %w", err)
			}

			// Set blockers: this task blocks the listed task IDs
			if blocks != "" {
				blockerIDs, err := parseBlocksFlag(blocks)
				if err != nil {
					return err
				}
				for _, bid := range blockerIDs {
					if err := c.taskStore.SetBlocker(bid, t.ID); err != nil {
						return fmt.Errorf("set blocker: %w", err)
					}
				}
			}

			if recur != "" && strings.ToLower(recur) != "none" {
				freq := task.ParseRecurFreq(strings.ToLower(recur))
				if freq == task.RecurNone {
					return fmt.Errorf("invalid recurrence %q: must be daily, weekly, monthly, or yearly", recur)
				}
				if err := c.taskStore.UpdateRecurrence(t.ID, freq, 1); err != nil {
					return fmt.Errorf("set recurrence: %w", err)
				}
			}

			if c.quiet {
				fmt.Fprintf(os.Stdout, "%d\n", t.ID)
			} else {
				c.printer(os.Stdout).Success("Created task #%d: %s", t.ID, t.Title)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&priority, "priority", "low", "Priority: low, medium, high, urgent")
	cmd.Flags().StringVar(&due, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&desc, "desc", "", "Task description")
	cmd.Flags().StringVar(&recur, "recur", "", "Recurrence: daily, weekly, monthly, yearly")
	cmd.Flags().StringSliceVar(&meta, "meta", nil, "Metadata key=value (repeatable: --meta source=whatsapp --meta group=main)")
	cmd.Flags().StringVar(&blocks, "blocks", "", "Comma-separated task IDs this task blocks")

	return cmd
}

func (c *CLI) doneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done <task-id> [task-id...]",
		Short: "Mark one or more tasks as done",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				id, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid task ID %q: %w", arg, err)
				}

				t, err := c.getTaskOrNotFound(id)
				if err != nil {
					return err
				}

				// Spawn next occurrence before marking the current task done.
				if t.RecurFreq != task.RecurNone {
					nextDue := task.NextDueDate(*t)
					interval := t.RecurInterval
					if interval <= 0 {
						interval = 1
					}
					newTask := &task.Task{
						Title:         t.Title,
						Description:   t.Description,
						Priority:      t.Priority,
						DueDate:       &nextDue,
						Tags:          t.Tags,
						RecurFreq:     t.RecurFreq,
						RecurInterval: interval,
					}
					if err := c.taskStore.Create(newTask); err != nil {
						return fmt.Errorf("spawn recurring task: %w", err)
					}
					if err := c.taskStore.UpdateRecurrence(newTask.ID, newTask.RecurFreq, interval); err != nil {
						return fmt.Errorf("set recurrence on new task: %w", err)
					}
				}

				t.Status = task.Done
				if err := c.taskStore.Update(t); err != nil {
					return fmt.Errorf("update task: %w", err)
				}

				c.printer(os.Stdout).Success("Marked task #%d as done: %s", t.ID, t.Title)
			}
			return nil
		},
	}
}

func (c *CLI) showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <task-id>",
		Short: "Show full task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}

			t, err := c.getTaskOrNotFound(id)
			if err != nil {
				return err
			}

			switch strings.ToLower(c.format) {
			case "json":
				return c.printTaskJSON(os.Stdout, t)
			default:
				return c.printTaskDetail(c.printer(os.Stdout), t)
			}
		},
	}
}

func (c *CLI) editCmd() *cobra.Command {
	var title, desc, priority, due, tags, recur, blocks string
	var clearDue, clearBlocks bool
	var meta []string

	cmd := &cobra.Command{
		Use:   "edit <task-id> [flags]",
		Short: "Edit a task (only specified flags are updated)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}

			t, err := c.getTaskOrNotFound(id)
			if err != nil {
				return err
			}

			changed := false

			if cmd.Flags().Changed("title") {
				t.Title = title
				changed = true
			}
			if cmd.Flags().Changed("desc") {
				t.Description = desc
				changed = true
			}
			if cmd.Flags().Changed("priority") {
				prio, err := parsePriority(priority)
				if err != nil {
					return err
				}
				t.Priority = prio
				changed = true
			}
			if clearDue {
				t.DueDate = nil
				changed = true
			} else if cmd.Flags().Changed("due") {
				d, err := time.ParseInLocation(time.DateOnly, due, time.UTC)
				if err != nil {
					return fmt.Errorf("invalid due date %q: expected YYYY-MM-DD", due)
				}
				t.DueDate = &d
				changed = true
			}
			if cmd.Flags().Changed("tags") {
				t.Tags = nil
				for _, tag := range strings.Split(tags, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						t.Tags = append(t.Tags, tag)
					}
				}
				changed = true
			}
			if cmd.Flags().Changed("meta") {
				newMeta, err := parseMetaFlags(meta)
				if err != nil {
					return err
				}
				if t.Metadata == nil {
					t.Metadata = make(map[string]string)
				}
				for k, v := range newMeta {
					t.Metadata[k] = v
				}
				changed = true
			}

			blocksChanged := clearBlocks || cmd.Flags().Changed("blocks")
			recurChanged := cmd.Flags().Changed("recur")
			if !changed && !recurChanged && !blocksChanged {
				return fmt.Errorf("no changes specified: use --title, --desc, --priority, --due, --tags, --meta, --blocks, --recur, --clear-due, or --clear-blocks")
			}

			if changed {
				if err := c.taskStore.Update(t); err != nil {
					return fmt.Errorf("update task: %w", err)
				}
			}

			if clearBlocks {
				if err := c.taskStore.SetBlocksIDs(id, nil); err != nil {
					return fmt.Errorf("clear blocks: %w", err)
				}
			} else if cmd.Flags().Changed("blocks") {
				newBlocks, err := parseBlocksFlag(blocks)
				if err != nil {
					return err
				}
				if err := c.taskStore.SetBlocksIDs(id, newBlocks); err != nil {
					return fmt.Errorf("set blocks: %w", err)
				}
			}

			if recurChanged {
				freq := task.ParseRecurFreq(strings.ToLower(recur))
				interval := t.RecurInterval
				if interval <= 0 {
					interval = 1
				}
				if err := c.taskStore.UpdateRecurrence(id, freq, interval); err != nil {
					return fmt.Errorf("set recurrence: %w", err)
				}
			}

			c.printer(os.Stdout).Success("Updated task #%d: %s", t.ID, t.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&desc, "desc", "", "New description")
	cmd.Flags().StringVar(&priority, "priority", "", "New priority: low, medium, high, urgent")
	cmd.Flags().StringVar(&due, "due", "", "New due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&tags, "tags", "", "New comma-separated tags (replaces all existing tags)")
	cmd.Flags().StringVar(&recur, "recur", "", "Recurrence: none, daily, weekly, monthly, yearly")
	cmd.Flags().BoolVar(&clearDue, "clear-due", false, "Remove the due date")
	cmd.Flags().StringSliceVar(&meta, "meta", nil, "Metadata key=value (repeatable, merges with existing)")
	cmd.Flags().StringVar(&blocks, "blocks", "", "Comma-separated task IDs this task blocks (replaces existing)")
	cmd.Flags().BoolVar(&clearBlocks, "clear-blocks", false, "Remove all tasks this task blocks")

	return cmd
}

func (c *CLI) deleteCmd() *cobra.Command {
	var force, cascade bool

	cmd := &cobra.Command{
		Use:     "delete <task-id>",
		Short:   "Delete a task",
		Aliases: []string{"del", "rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}

			t, err := c.getTaskOrNotFound(id)
			if err != nil {
				return err
			}

			// Guard: refuse to delete if this task blocks others (unless --cascade)
			blocksIDs, err := c.taskStore.ListBlocksIDs(id)
			if err != nil {
				return fmt.Errorf("check dependencies: %w", err)
			}
			if len(blocksIDs) > 0 && !cascade {
				idStrs := make([]string, len(blocksIDs))
				for i, bid := range blocksIDs {
					idStrs[i] = fmt.Sprintf("#%d", bid)
				}
				return fmt.Errorf("task #%d blocks %s. Use --cascade to delete and unblock them", id, strings.Join(idStrs, ", "))
			}

			ok, err := Confirm(fmt.Sprintf("Delete task #%d %q?", id, t.Title), force)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}

			if err := c.taskStore.Delete(id); err != nil {
				return fmt.Errorf("delete task: %w", err)
			}

			if len(blocksIDs) > 0 {
				idStrs := make([]string, len(blocksIDs))
				for i, bid := range blocksIDs {
					idStrs[i] = fmt.Sprintf("#%d", bid)
				}
				c.printer(os.Stdout).Success("Deleted task #%d: %s (unblocked %s)", id, t.Title, strings.Join(idStrs, ", "))
			} else {
				c.printer(os.Stdout).Success("Deleted task #%d: %s", id, t.Title)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&cascade, "cascade", false, "Delete even if this task blocks others (unblocks them)")

	return cmd
}

func (c *CLI) statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <task-id> [pending|active|done]",
		Short: "Set or cycle task status",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}

			t, err := c.getTaskOrNotFound(id)
			if err != nil {
				return err
			}

			if len(args) == 2 {
				switch strings.ToLower(args[1]) {
				case "pending":
					t.Status = task.Pending
				case "active", "in-progress", "inprogress":
					t.Status = task.InProgress
				case "done":
					t.Status = task.Done
				default:
					return fmt.Errorf("invalid status %q: must be pending, active, or done", args[1])
				}
			} else {
				t.Status = t.Status.Next()
			}

			if err := c.taskStore.Update(t); err != nil {
				return fmt.Errorf("update task: %w", err)
			}

			c.printer(os.Stdout).Success("Task #%d status: %s", t.ID, t.Status.String())
			return nil
		},
	}
}

func (c *CLI) listCmd() *cobra.Command {
	var status, priority, sortBy, dueBefore, dueAfter, search string
	var tags []string
	var metaFilter []string
	var overdue bool
	var limit int

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := c.taskStore.List()
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}

			filtered, err := applyTaskFilters(tasks, taskFilterOpts{
				status:     status,
				priority:   priority,
				tags:       tags,
				metaFilter: metaFilter,
				dueBefore:  dueBefore,
				dueAfter:   dueAfter,
				overdue:    overdue,
				search:     search,
			})
			if err != nil {
				return err
			}

			sortTasks(filtered, sortBy)

			if limit > 0 && len(filtered) > limit {
				filtered = filtered[:limit]
			}

			switch strings.ToLower(c.format) {
			case "json":
				return printTasksJSON(os.Stdout, filtered)
			default:
				return c.printTasksTable(c.printer(os.Stdout), filtered)
			}
		},
	}

	cmd.Flags().StringVar(&status, "status", "all", "Filter: pending, active, done, all")
	cmd.Flags().StringVar(&priority, "priority", "", "Filter by priority: low, medium, high, urgent")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Filter by tag (repeatable: --tag work --tag urgent)")
	cmd.Flags().StringSliceVar(&metaFilter, "meta", nil, "Filter by metadata key=value (repeatable, AND logic)")
	cmd.Flags().StringVar(&sortBy, "sort", "created", "Sort order: created, due, priority")
	cmd.Flags().StringVar(&dueBefore, "due-before", "", "Show tasks due on or before YYYY-MM-DD")
	cmd.Flags().StringVar(&dueAfter, "due-after", "", "Show tasks due on or after YYYY-MM-DD")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "Show only overdue tasks (not done, past due date)")
	cmd.Flags().StringVar(&search, "search", "", "Substring match on title and description")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of tasks to show (0 = unlimited)")

	return cmd
}

// taskFilterOpts holds all list filter parameters.
type taskFilterOpts struct {
	status     string
	priority   string
	tags       []string
	metaFilter []string
	dueBefore  string
	dueAfter   string
	overdue    bool
	search     string
}

func applyTaskFilters(tasks []task.Task, opts taskFilterOpts) ([]task.Task, error) {
	tasks = filterTasks(tasks, opts.status)

	if opts.priority != "" {
		prio, err := parsePriority(opts.priority)
		if err != nil {
			return nil, err
		}
		tasks = filterSlice(tasks, func(t task.Task) bool {
			return t.Priority == prio
		})
	}

	if len(opts.tags) > 0 {
		tasks = filterSlice(tasks, func(t task.Task) bool {
			return taskHasAnyTag(t, opts.tags)
		})
	}

	if len(opts.metaFilter) > 0 {
		metaKV, err := parseMetaFlags(opts.metaFilter)
		if err != nil {
			return nil, err
		}
		tasks = filterSlice(tasks, func(t task.Task) bool {
			return taskMatchesMeta(t, metaKV)
		})
	}

	if opts.search != "" {
		lower := strings.ToLower(opts.search)
		tasks = filterSlice(tasks, func(t task.Task) bool {
			return strings.Contains(strings.ToLower(t.Title), lower) ||
				strings.Contains(strings.ToLower(t.Description), lower)
		})
	}

	if opts.overdue {
		now := time.Now()
		tasks = filterSlice(tasks, func(t task.Task) bool {
			return t.DueDate != nil && t.DueDate.Before(now) && t.Status != task.Done
		})
	}

	if opts.dueBefore != "" {
		if cutoff, err := time.ParseInLocation(time.DateOnly, opts.dueBefore, time.UTC); err == nil {
			tasks = filterSlice(tasks, func(t task.Task) bool {
				return t.DueDate != nil && !t.DueDate.After(cutoff)
			})
		}
	}

	if opts.dueAfter != "" {
		if cutoff, err := time.ParseInLocation(time.DateOnly, opts.dueAfter, time.UTC); err == nil {
			tasks = filterSlice(tasks, func(t task.Task) bool {
				return t.DueDate != nil && !t.DueDate.Before(cutoff)
			})
		}
	}

	return tasks, nil
}

// filterSlice returns tasks for which pred returns true.
func filterSlice(tasks []task.Task, pred func(task.Task) bool) []task.Task {
	var out []task.Task
	for _, t := range tasks {
		if pred(t) {
			out = append(out, t)
		}
	}
	return out
}

// taskHasAnyTag reports whether the task has at least one of the given tags.
func taskHasAnyTag(t task.Task, filterTags []string) bool {
	tagSet := make(map[string]struct{}, len(t.Tags))
	for _, tag := range t.Tags {
		tagSet[strings.ToLower(tag)] = struct{}{}
	}
	for _, ft := range filterTags {
		if _, ok := tagSet[strings.ToLower(ft)]; ok {
			return true
		}
	}
	return false
}

// taskMatchesMeta reports whether the task matches all given metadata key=value pairs (AND logic).
func taskMatchesMeta(t task.Task, filter map[string]string) bool {
	if t.Metadata == nil {
		return false
	}
	for k, v := range filter {
		if t.Metadata[k] != v {
			return false
		}
	}
	return true
}

// sortTasks sorts tasks in-place by the given sort key.
func sortTasks(tasks []task.Task, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "due":
		sort.Slice(tasks, func(i, j int) bool {
			di, dj := tasks[i].DueDate, tasks[j].DueDate
			if di == nil {
				return false // nil due dates go last
			}
			if dj == nil {
				return true
			}
			return di.Before(*dj)
		})
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority > tasks[j].Priority // higher priority first
		})
	default: // "created" — most recent first
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
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

func (c *CLI) printTasksTable(p *Printer, tasks []task.Task) error {
	rows := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		due := "-"
		if t.DueDate != nil {
			due = c.cfg.FormatDate(*t.DueDate)
		}
		tags := "-"
		if len(t.Tags) > 0 {
			tags = strings.Join(t.Tags, ", ")
		}
		completedSubs := 0
		for _, s := range t.Subtasks {
			if s.Completed {
				completedSubs++
			}
		}
		subs := "-"
		if len(t.Subtasks) > 0 {
			subs = fmt.Sprintf("%d/%d", completedSubs, len(t.Subtasks))
		}
		notes := "-"
		if len(t.Notes) > 0 {
			notes = fmt.Sprintf("%d", len(t.Notes))
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", t.ID),
			formatStatus(p, t.Status),
			formatPriority(p, t.Priority),
			t.Title,
			due,
			tags,
			subs,
			notes,
		})
	}
	p.Table([]string{"ID", "STATUS", "PRIORITY", "TITLE", "DUE", "TAGS", "SUBS", "NOTES"}, rows)
	return nil
}

func formatStatus(p *Printer, s task.Status) string {
	switch s {
	case task.InProgress:
		return p.Colored("● Active", ui.Cyan)
	case task.Done:
		return p.Colored("✓ Done", ui.Green)
	default:
		return p.Colored("○ Pending", ui.Gray)
	}
}

func formatPriority(p *Printer, pr task.Priority) string {
	switch pr {
	case task.Urgent:
		return p.Colored("Urgent", ui.Red)
	case task.High:
		return p.Colored("High", ui.Orange)
	case task.Medium:
		return p.Colored("Medium", ui.Yellow)
	default:
		return p.Colored("Low", ui.Gray)
	}
}

func (c *CLI) printTaskDetail(p *Printer, t *task.Task) error {
	w := p.w

	due := "-"
	if t.DueDate != nil {
		due = c.cfg.FormatDate(*t.DueDate)
	}
	tags := "-"
	if len(t.Tags) > 0 {
		tags = strings.Join(t.Tags, ", ")
	}
	completedSubs := 0
	for _, s := range t.Subtasks {
		if s.Completed {
			completedSubs++
		}
	}
	subtaskStr := fmt.Sprintf("%d/%d", completedSubs, len(t.Subtasks))

	var blockedByStrs []string
	for _, bid := range t.BlockedByIDs {
		blockedByStrs = append(blockedByStrs, fmt.Sprintf("#%d", bid))
	}
	blockedBy := "-"
	if len(blockedByStrs) > 0 {
		blockedBy = strings.Join(blockedByStrs, ", ")
	}

	var blocksStrs []string
	for _, bid := range t.BlocksIDs {
		blocksStrs = append(blocksStrs, fmt.Sprintf("#%d", bid))
	}
	blocks := "-"
	if len(blocksStrs) > 0 {
		blocks = strings.Join(blocksStrs, ", ")
	}

	meta := "-"
	if len(t.Metadata) > 0 {
		keys := make([]string, 0, len(t.Metadata))
		for k := range t.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var pairs []string
		for _, k := range keys {
			pairs = append(pairs, k+"="+t.Metadata[k])
		}
		meta = strings.Join(pairs, ", ")
	}

	label := func(s string) string { return p.Bold(s) }
	status := formatStatus(p, t.Status)
	priority := formatPriority(p, t.Priority)

	p.Table(
		[]string{"FIELD", "VALUE"},
		[][]string{
			{label("ID"), fmt.Sprintf("%d", t.ID)},
			{label("Title"), t.Title},
			{label("Status"), status},
			{label("Priority"), priority},
			{label("Due Date"), due},
			{label("Tags"), tags},
			{label("Metadata"), meta},
			{label("Recurrence"), t.RecurFreq.String()},
			{label("Subtasks"), subtaskStr},
			{label("Blocked By"), blockedBy},
			{label("Blocks"), blocks},
			{label("Notes"), fmt.Sprintf("%d", len(t.Notes))},
			{label("Time Logged"), task.FormatDuration(task.TotalDuration(t.TimeLogs))},
			{label("Created"), c.cfg.FormatDateTime(t.CreatedAt)},
			{label("Updated"), c.cfg.FormatDateTime(t.UpdatedAt)},
		},
	)

	if t.Description != "" {
		fmt.Fprintf(w, "\n%s\n  %s\n", p.Bold("Description:"), t.Description)
	}
	if len(t.Subtasks) > 0 {
		fmt.Fprintf(w, "\n%s\n", p.Bold("Subtasks:"))
		for _, s := range t.Subtasks {
			check := p.Colored("○", ui.Gray)
			if s.Completed {
				check = p.Colored("✓", ui.Green)
			}
			fmt.Fprintf(w, "  %s [%d] %s\n", check, s.ID, s.Title)
		}
	}
	if len(t.TimeLogs) > 0 {
		fmt.Fprintf(w, "\n%s\n", p.Bold("Time Logs:"))
		for _, tl := range t.TimeLogs {
			note := ""
			if tl.Note != "" {
				note = " — " + tl.Note
			}
			fmt.Fprintf(w, "  %s  [%d] %s%s\n",
				p.Dim(c.cfg.FormatDateTime(tl.LoggedAt)),
				tl.ID,
				task.FormatDuration(tl.Duration),
				note,
			)
		}
	}
	if len(t.Notes) > 0 {
		fmt.Fprintf(w, "\n%s\n", p.Bold("Notes:"))
		for _, n := range t.Notes {
			fmt.Fprintf(w, "  %s  [%d] %s\n",
				p.Dim(c.cfg.FormatDateTime(n.CreatedAt)),
				n.ID,
				n.Body,
			)
		}
	}
	return nil
}

// JSON output types

type jsonSubtask struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Position  int    `json:"position"`
}

type jsonTimeLog struct {
	ID       int64  `json:"id"`
	Duration string `json:"duration"`
	Note     string `json:"note,omitempty"`
	LoggedAt string `json:"logged_at"`
}

type jsonTaskNote struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type jsonTaskRef struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type jsonTask struct {
	ID              int64             `json:"id"`
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	Status          string            `json:"status"`
	Priority        string            `json:"priority"`
	DueDate         string            `json:"due_date,omitempty"`
	Tags            []string          `json:"tags"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Recurrence      string            `json:"recurrence,omitempty"`
	Subtasks        []jsonSubtask     `json:"subtasks"`
	TimeLogs        []jsonTimeLog     `json:"time_logs"`
	NoteCount       int               `json:"note_count"`
	Notes           []jsonTaskNote    `json:"notes"`
	BlockedBy       []int64           `json:"blocked_by"`
	BlockedByDetail []jsonTaskRef     `json:"blocked_by_detail"`
	Blocks          []int64           `json:"blocks"`
	BlocksDetail    []jsonTaskRef     `json:"blocks_detail"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}

// taskIndex maps task IDs to their summary info for enriching dependency refs.
type taskIndex map[int64]jsonTaskRef

func buildTaskIndex(tasks []task.Task) taskIndex {
	idx := make(taskIndex, len(tasks))
	for _, t := range tasks {
		idx[t.ID] = jsonTaskRef{ID: t.ID, Title: t.Title, Status: t.Status.String()}
	}
	return idx
}

func resolveRefs(ids []int64, idx taskIndex) []jsonTaskRef {
	refs := make([]jsonTaskRef, 0, len(ids))
	for _, id := range ids {
		if ref, ok := idx[id]; ok {
			refs = append(refs, ref)
		} else {
			refs = append(refs, jsonTaskRef{ID: id, Title: "(unknown)", Status: "Unknown"})
		}
	}
	return refs
}

func taskToJSON(t *task.Task, idx taskIndex) jsonTask {
	jt := jsonTask{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status.String(),
		Priority:    t.Priority.String(),
		Tags:        t.Tags,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
	if jt.Tags == nil {
		jt.Tags = []string{}
	}
	jt.Metadata = t.Metadata
	if t.DueDate != nil {
		jt.DueDate = t.DueDate.Format(time.DateOnly)
	}
	if t.RecurFreq != task.RecurNone {
		jt.Recurrence = t.RecurFreq.String()
	}
	jt.Subtasks = make([]jsonSubtask, 0, len(t.Subtasks))
	for _, s := range t.Subtasks {
		jt.Subtasks = append(jt.Subtasks, jsonSubtask{
			ID:        s.ID,
			Title:     s.Title,
			Completed: s.Completed,
			Position:  s.Position,
		})
	}
	jt.TimeLogs = make([]jsonTimeLog, 0, len(t.TimeLogs))
	for _, tl := range t.TimeLogs {
		jt.TimeLogs = append(jt.TimeLogs, jsonTimeLog{
			ID:       tl.ID,
			Duration: task.FormatDuration(tl.Duration),
			Note:     tl.Note,
			LoggedAt: tl.LoggedAt.Format(time.RFC3339),
		})
	}
	jt.NoteCount = len(t.Notes)
	jt.Notes = make([]jsonTaskNote, 0, len(t.Notes))
	for _, n := range t.Notes {
		jt.Notes = append(jt.Notes, jsonTaskNote{
			ID:        n.ID,
			Body:      n.Body,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		})
	}
	jt.BlockedBy = t.BlockedByIDs
	if jt.BlockedBy == nil {
		jt.BlockedBy = []int64{}
	}
	jt.BlockedByDetail = resolveRefs(t.BlockedByIDs, idx)
	jt.Blocks = t.BlocksIDs
	if jt.Blocks == nil {
		jt.Blocks = []int64{}
	}
	jt.BlocksDetail = resolveRefs(t.BlocksIDs, idx)
	return jt
}

func printTasksJSON(w io.Writer, tasks []task.Task) error {
	idx := buildTaskIndex(tasks)
	out := make([]jsonTask, 0, len(tasks))
	for i := range tasks {
		out = append(out, taskToJSON(&tasks[i], idx))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func (c *CLI) printTaskJSON(w io.Writer, t *task.Task) error {
	// Build index from only referenced tasks (not all tasks).
	idx := taskIndex{t.ID: {ID: t.ID, Title: t.Title, Status: t.Status.String()}}
	refIDs := make(map[int64]bool)
	for _, id := range t.BlockedByIDs {
		refIDs[id] = true
	}
	for _, id := range t.BlocksIDs {
		refIDs[id] = true
	}
	for id := range refIDs {
		if id == t.ID {
			continue
		}
		ref, err := c.taskStore.GetByID(id)
		if err == nil {
			idx[ref.ID] = jsonTaskRef{ID: ref.ID, Title: ref.Title, Status: ref.Status.String()}
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(taskToJSON(t, idx))
}
