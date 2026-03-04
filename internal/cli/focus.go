package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/roniel/todo-app/internal/focus"
	"github.com/roniel/todo-app/internal/task"
	"github.com/spf13/cobra"
)

func (c *CLI) focusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Manage focus (Pomodoro) sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(c.focusStartCmd())
	cmd.AddCommand(c.focusStatusCmd())
	cmd.AddCommand(c.focusStatsCmd())

	return cmd
}

func (c *CLI) focusStartCmd() *cobra.Command {
	var taskID int64
	var durationStr string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Record a completed focus session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			defaultDur := fmt.Sprintf("%dm", c.cfg.Focus.WorkDuration)
			if !cmd.Flags().Changed("duration") {
				durationStr = defaultDur
			}

			dur, err := task.ParseDuration(durationStr)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}

			now := time.Now()
			sess := &focus.Session{
				TaskID:    taskID,
				Duration:  dur,
				StartedAt: now,
				Kind:      focus.KindWork,
				CyclePos:  1,
			}

			if err := c.focusStore.Create(sess); err != nil {
				return fmt.Errorf("create session: %w", err)
			}

			// Immediately complete the session (CLI has no interactive timer).
			if err := c.focusStore.Complete(sess.ID); err != nil {
				return fmt.Errorf("complete session: %w", err)
			}

			p := c.printer(os.Stdout)
			if c.quiet {
				fmt.Fprintf(os.Stdout, "%d\n", sess.ID)
			} else {
				p.Success("Recorded focus session #%d (%s)", sess.ID, task.FormatDuration(dur))
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&taskID, "task-id", 0, "Associate session with a task ID")
	cmd.Flags().StringVar(&durationStr, "duration", "", "Session duration (e.g. 25m, 1h); default from config")

	return cmd
}

func (c *CLI) focusStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show today's focus status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			todayWork, err := c.focusStore.TodayWorkCount()
			if err != nil {
				return fmt.Errorf("today work count: %w", err)
			}
			streak, err := c.focusStore.Streak()
			if err != nil {
				return fmt.Errorf("streak: %w", err)
			}
			goal := c.cfg.Focus.DailyGoal

			switch strings.ToLower(c.format) {
			case "json":
				out := map[string]any{
					"today":       todayWork,
					"goal":        goal,
					"streak_days": streak,
					"date":        time.Now().Format(time.DateOnly),
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			default:
				p := c.printer(os.Stdout)
				p.Table(
					[]string{"METRIC", "VALUE"},
					[][]string{
						{"Today", fmt.Sprintf("%d / %d (goal)", todayWork, goal)},
						{"Streak", fmt.Sprintf("%d days", streak)},
						{"Date", c.cfg.FormatDate(time.Now())},
					},
				)
				return nil
			}
		},
	}
}

func (c *CLI) focusStatsCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show focus session statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			byDay, err := c.focusStore.CompletionsByDay(days)
			if err != nil {
				return fmt.Errorf("completions by day: %w", err)
			}

			// Sort dates descending.
			dates := make([]string, 0, len(byDay))
			for d := range byDay {
				dates = append(dates, d)
			}
			sort.Slice(dates, func(i, j int) bool {
				return dates[i] > dates[j]
			})

			switch strings.ToLower(c.format) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(byDay)
			default:
				rows := make([][]string, 0, len(dates))
				for _, d := range dates {
					rows = append(rows, []string{d, fmt.Sprintf("%d", byDay[d])})
				}
				if len(rows) == 0 {
					p := c.printer(os.Stdout)
					fmt.Fprintln(p.w, p.Dim(fmt.Sprintf("(no focus sessions in the last %d days)", days)))
					return nil
				}
				p := c.printer(os.Stdout)
				p.Table([]string{"DATE", "SESSIONS"}, rows)
				return nil
			}
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "Number of days to show")

	return cmd
}
