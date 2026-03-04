package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/roniel/todo-app/internal/task"
	"github.com/spf13/cobra"
)

func (c *CLI) statsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show a summary of tasks and focus sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := c.taskStore.List()
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}

			// Count by status.
			var pending, active, done int
			for _, t := range tasks {
				switch t.Status {
				case task.Pending:
					pending++
				case task.InProgress:
					active++
				case task.Done:
					done++
				}
			}

			// Count by priority (non-done tasks only).
			priCounts := map[string]int{"Low": 0, "Medium": 0, "High": 0, "Urgent": 0}
			for _, t := range tasks {
				if t.Status != task.Done {
					priCounts[t.Priority.String()]++
				}
			}

			// Focus stats.
			todayWork, err := c.focusStore.TodayWorkCount()
			if err != nil {
				return fmt.Errorf("focus today count: %w", err)
			}
			streak, err := c.focusStore.Streak()
			if err != nil {
				return fmt.Errorf("focus streak: %w", err)
			}
			totalMin, err := c.focusStore.TotalMinutesFocused(30)
			if err != nil {
				return fmt.Errorf("total minutes focused: %w", err)
			}

			goal := c.cfg.Focus.DailyGoal

			switch strings.ToLower(c.format) {
			case "json":
				return c.printer(os.Stdout).JSON(map[string]any{
					"tasks": map[string]any{
						"total":   len(tasks),
						"pending": pending,
						"active":  active,
						"done":    done,
						"by_priority": map[string]int{
							"low":    priCounts["Low"],
							"medium": priCounts["Medium"],
							"high":   priCounts["High"],
							"urgent": priCounts["Urgent"],
						},
					},
					"focus": map[string]any{
						"today":            todayWork,
						"goal":             goal,
						"streak_days":      streak,
						"total_min_30days": totalMin,
					},
				})
			default:
				return c.printStatsTable(c.printer(os.Stdout), pending, active, done, priCounts, todayWork, goal, streak, totalMin)
			}
		},
	}
}

func (c *CLI) printStatsTable(p *Printer, pending, active, done int, priCounts map[string]int,
	todayWork, goal, streak, totalMin int) error {

	fmt.Fprintln(p.w, p.Bold("TASKS"))
	p.Table(
		[]string{"STATUS", "COUNT"},
		[][]string{
			{"Pending", fmt.Sprintf("%d", pending)},
			{"Active", fmt.Sprintf("%d", active)},
			{"Done", fmt.Sprintf("%d", done)},
			{"Total", fmt.Sprintf("%d", pending+active+done)},
		},
	)

	fmt.Fprintln(p.w)
	fmt.Fprintln(p.w, p.Bold("OPEN TASKS BY PRIORITY"))
	p.Table(
		[]string{"PRIORITY", "COUNT"},
		[][]string{
			{"Urgent", fmt.Sprintf("%d", priCounts["Urgent"])},
			{"High", fmt.Sprintf("%d", priCounts["High"])},
			{"Medium", fmt.Sprintf("%d", priCounts["Medium"])},
			{"Low", fmt.Sprintf("%d", priCounts["Low"])},
		},
	)

	fmt.Fprintln(p.w)
	fmt.Fprintln(p.w, p.Bold("FOCUS (last 30 days)"))
	totalHours := totalMin / 60
	totalMinsRem := totalMin % 60
	totalFmt := fmt.Sprintf("%dh %dm", totalHours, totalMinsRem)
	if totalHours == 0 {
		totalFmt = fmt.Sprintf("%dm", totalMinsRem)
	}
	p.Table(
		[]string{"METRIC", "VALUE"},
		[][]string{
			{"Today", fmt.Sprintf("%d / %d (goal)", todayWork, goal)},
			{"Streak", fmt.Sprintf("%d days", streak)},
			{"Total focused", totalFmt},
			{"As of", c.cfg.FormatDate(time.Now())},
		},
	)

	return nil
}
