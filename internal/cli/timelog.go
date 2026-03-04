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

func (c *CLI) timelogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timelog",
		Short: "Manage time logs",
	}
	cmd.AddCommand(
		c.timelogAddCmd(),
		c.timelogListCmd(),
		c.timelogSummaryCmd(),
	)
	return cmd
}

func (c *CLI) timelogAddCmd() *cobra.Command {
	var note string
	cmd := &cobra.Command{
		Use:   "add <task-id> <duration>",
		Short: "Log time spent on a task (e.g. 1h30m, 45m, 2h)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}
			if _, err := c.getTaskOrNotFound(taskID); err != nil {
				return err
			}
			dur, err := task.ParseDuration(args[1])
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			if err := c.taskStore.AddTimeLog(taskID, dur, note); err != nil {
				return fmt.Errorf("add time log: %w", err)
			}
			c.printer(os.Stdout).Success("Logged %s to task #%d", task.FormatDuration(dur), taskID)
			return nil
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "Optional note for this time entry")
	return cmd
}

func (c *CLI) timelogListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <task-id>",
		Short: "List time logs for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID %q: %w", args[0], err)
			}
			// Verify task exists.
			if _, err := c.getTaskOrNotFound(taskID); err != nil {
				return err
			}
			logs, err := c.taskStore.ListTimeLogs(taskID)
			if err != nil {
				return fmt.Errorf("list time logs: %w", err)
			}
			p := c.printer(os.Stdout)
			if strings.ToLower(c.format) == "json" {
				type jsonTL struct {
					ID       int64  `json:"id"`
					Date     string `json:"date"`
					Duration string `json:"duration"`
					Note     string `json:"note,omitempty"`
				}
				out := make([]jsonTL, 0, len(logs))
				for _, l := range logs {
					out = append(out, jsonTL{
						ID:       l.ID,
						Date:     l.LoggedAt.Format("2006-01-02"),
						Duration: task.FormatDuration(l.Duration),
						Note:     l.Note,
					})
				}
				return p.JSON(out)
			}
			rows := make([][]string, 0, len(logs))
			for _, l := range logs {
				rows = append(rows, []string{
					c.cfg.FormatDate(l.LoggedAt),
					task.FormatDuration(l.Duration),
					l.Note,
				})
			}
			p.Table([]string{"DATE", "DURATION", "NOTE"}, rows)
			total := task.TotalDuration(logs)
			p.Success("Total: %s", task.FormatDuration(total))
			return nil
		},
	}
}

func (c *CLI) timelogSummaryCmd() *cobra.Command {
	var days int
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Summarize time logged across all tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := c.taskStore.List()
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}

			cutoff := time.Now().AddDate(0, 0, -days)

			type taskSummary struct {
				taskID    int64
				taskTitle string
				total     time.Duration
			}

			var summaries []taskSummary
			var grandTotal time.Duration

			for _, t := range tasks {
				var taskTotal time.Duration
				for _, l := range t.TimeLogs {
					if l.LoggedAt.After(cutoff) {
						taskTotal += l.Duration
					}
				}
				if taskTotal > 0 {
					summaries = append(summaries, taskSummary{
						taskID:    t.ID,
						taskTitle: t.Title,
						total:     taskTotal,
					})
					grandTotal += taskTotal
				}
			}

			p := c.printer(os.Stdout)
			if strings.ToLower(c.format) == "json" {
				type jsonSummary struct {
					TaskID    int64  `json:"task_id"`
					TaskTitle string `json:"task_title"`
					Duration  string `json:"duration"`
				}
				out := make([]jsonSummary, 0, len(summaries))
				for _, s := range summaries {
					out = append(out, jsonSummary{
						TaskID:    s.taskID,
						TaskTitle: s.taskTitle,
						Duration:  task.FormatDuration(s.total),
					})
				}
				return p.JSON(out)
			}
			rows := make([][]string, 0, len(summaries))
			for _, s := range summaries {
				rows = append(rows, []string{
					strconv.FormatInt(s.taskID, 10),
					s.taskTitle,
					task.FormatDuration(s.total),
				})
			}
			p.Table([]string{"ID", "TASK", "TOTAL"}, rows)
			p.Success("Grand total (%d days): %s", days, task.FormatDuration(grandTotal))
			return nil
		},
	}
	cmd.Flags().IntVar(&days, "days", 7, "Number of past days to include")
	return cmd
}
