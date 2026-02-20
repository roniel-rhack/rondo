// internal/app/delegate.go
package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
)

type taskDelegate struct{}

func newTaskDelegate() taskDelegate {
	return taskDelegate{}
}

func (d taskDelegate) Height() int  { return 2 }
func (d taskDelegate) Spacing() int { return 0 }

func (d taskDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	t, ok := item.(task.Task)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Status icon
	var statusColor lipgloss.Color
	switch t.Status {
	case task.InProgress:
		statusColor = ui.Yellow
	case task.Done:
		statusColor = ui.Green
	default:
		statusColor = ui.Gray
	}
	statusIcon := lipgloss.NewStyle().Foreground(statusColor).Render(t.Status.Icon())

	// Priority label
	var prioColor lipgloss.Color
	switch t.Priority {
	case task.Urgent:
		prioColor = ui.Magenta
	case task.High:
		prioColor = ui.Red
	case task.Medium:
		prioColor = ui.Yellow
	default:
		prioColor = ui.Green
	}
	prioLabel := lipgloss.NewStyle().Foreground(prioColor).Render(t.Priority.Label())

	// Recurring icon
	recurIcon := ""
	if t.RecurFreq != task.RecurNone {
		recurIcon = lipgloss.NewStyle().Foreground(ui.Cyan).Render(" ↻")
	}

	// Blocked badge
	blockedBadge := ""
	if len(t.BlockedByIDs) > 0 {
		blockedBadge = lipgloss.NewStyle().Foreground(ui.Red).Bold(true).Render(" [BLOCKED]")
	}

	// Title line
	titleStyle := lipgloss.NewStyle().Foreground(ui.White)
	if t.Status == task.Done {
		titleStyle = titleStyle.Strikethrough(true).Foreground(ui.Gray)
	}

	line1 := fmt.Sprintf(" %s %s %s%s%s", statusIcon, prioLabel, titleStyle.Render(t.Title), recurIcon, blockedBadge)

	// Subtitle line: due date (with overdue styling) + subtask count
	var subtitle string
	if t.DueDate != nil {
		level := ui.DueStatus(*t.DueDate)
		dueStr := fmt.Sprintf("due %s", t.DueDate.Format("Jan 02"))
		badge := ui.DueBadge(level)
		if badge != "" {
			dueStr += " " + badge
		}
		subtitle += ui.DueStyle(level).Render(dueStr)
	}
	if len(t.Subtasks) > 0 {
		done := 0
		for _, st := range t.Subtasks {
			if st.Completed {
				done++
			}
		}
		if subtitle != "" {
			subtitle += "  "
		}
		subtitle += lipgloss.NewStyle().Foreground(ui.Gray).Render(fmt.Sprintf("[%d/%d]", done, len(t.Subtasks)))
	}
	line2 := lipgloss.NewStyle().PaddingLeft(5).Render(subtitle)

	// Cursor / selection
	if isSelected {
		cursor := lipgloss.NewStyle().Foreground(ui.Cyan).Render("▸")
		line1 = cursor + strings.TrimPrefix(line1, " ")
		line1 = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Render(line1)
		line2 = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")).Render(line2)
	}

	fmt.Fprintf(w, "%s\n%s", line1, line2)
}
