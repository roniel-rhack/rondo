package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
)

func (m Model) renderStatsOverlay() string {
	if m.stats == nil {
		return ""
	}
	s := m.stats

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Statistics"))
	lines = append(lines, "")

	// Task counts.
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Tasks"))
	lines = append(lines, fmt.Sprintf("  Total: %d  Active: %d  Done: %d", s.totalTasks, s.activeTasks, s.doneTasks))
	lines = append(lines, "")

	// Priority breakdown.
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Priority"))
	lines = append(lines, "  "+ui.RenderPriorityBreakdown(s.lowCount, s.medCount, s.highCount, s.urgentCount))
	lines = append(lines, "")

	// Tags.
	if len(s.tagCounts) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Tags"))
		lines = append(lines, "  "+ui.RenderTagCloud(s.tagCounts))
		lines = append(lines, "")
	}

	// Focus sessions today.
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Focus"))
	lines = append(lines, fmt.Sprintf("  Sessions today: %d", s.focusToday))
	lines = append(lines, "")

	// Journal streak.
	if s.journalStreak != "" {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Journal"))
		lines = append(lines, "  "+s.journalStreak)
		lines = append(lines, "")
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc or G to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 3).
		Width(60).
		Render(content)
}

func (m Model) renderBlockerOverlay() string {
	selected := m.selectedTask()
	if selected == nil {
		return ""
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Task Dependencies"))
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("Task: %s", selected.Title)))
	lines = append(lines, "")

	if len(selected.BlockedByIDs) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("No blockers"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Blocked by:"))
		for _, id := range selected.BlockedByIDs {
			blocker, err := m.store.GetByID(id)
			if err != nil {
				lines = append(lines, fmt.Sprintf("  - Task #%d (not found)", id))
				continue
			}
			statusIcon := blocker.Status.Icon()
			style := lipgloss.NewStyle().Foreground(ui.White)
			if blocker.Status == task.Done {
				style = style.Foreground(ui.Green)
			}
			lines = append(lines, style.Render(fmt.Sprintf("  %s %s", statusIcon, blocker.Title)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 2).
		Width(50).
		Render(content)
}

func (m *Model) computeStats() {
	s := &statsData{
		totalTasks: len(m.tasks),
		tagCounts:  make(map[string]int),
	}
	for _, t := range m.tasks {
		switch t.Status {
		case task.Done:
			s.doneTasks++
		default:
			s.activeTasks++
		}
		switch t.Priority {
		case task.Low:
			s.lowCount++
		case task.Medium:
			s.medCount++
		case task.High:
			s.highCount++
		case task.Urgent:
			s.urgentCount++
		}
		for _, tag := range t.Tags {
			s.tagCounts[tag]++
		}
	}
	if m.focusStore != nil {
		s.focusToday, _ = m.focusStore.TodayCount()
	}
	if m.journalStore != nil {
		completions := make(map[string]int)
		for _, n := range m.notes {
			if len(n.Entries) > 0 {
				completions[n.Date.Format(time.DateOnly)] = len(n.Entries)
			}
		}
		s.journalStreak = ui.RenderJournalStreak(completions, 30)
	}
	m.stats = s
}

// renderPanel draws a bordered panel with the title embedded in the top border (lazygit-style).
func renderPanel(content, title string, width, height int, focused bool) string {
	if width < 4 || height < 3 {
		return content
	}

	borderColor := gray
	if focused {
		borderColor = cyan
	}

	bc := lipgloss.NewStyle().Foreground(borderColor)
	tc := lipgloss.NewStyle().Foreground(borderColor).Bold(focused)

	border := lipgloss.RoundedBorder()
	innerWidth := width - 2

	// Pad and size the content area.
	padded := lipgloss.NewStyle().
		Width(innerWidth).
		Height(height - 2).
		PaddingLeft(1).
		PaddingRight(1).
		Render(content)

	// Top border with title: ╭─ 1 Tasks ──────────────╮
	titleRendered := tc.Render(title)
	titleVisualWidth := lipgloss.Width(titleRendered)
	fillWidth := innerWidth - titleVisualWidth - 3 // "─ " before + " " after
	if fillWidth < 0 {
		fillWidth = 0
	}
	topLine := bc.Render(border.TopLeft+border.Top+" ") +
		titleRendered +
		bc.Render(" "+strings.Repeat(border.Top, fillWidth)+border.TopRight)

	// Bottom border: ╰──────────────╯
	bottomLine := bc.Render(border.BottomLeft + strings.Repeat(border.Bottom, innerWidth) + border.BottomRight)

	// Add side borders to each content line.
	lines := strings.Split(padded, "\n")
	result := make([]string, 0, len(lines)+2)
	result = append(result, topLine)
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > innerWidth {
			line = lipgloss.NewStyle().MaxWidth(innerWidth).Render(line)
			lineWidth = lipgloss.Width(line)
		}
		if lineWidth < innerWidth {
			line += strings.Repeat(" ", innerWidth-lineWidth)
		}
		result = append(result, bc.Render(border.Left)+line+bc.Render(border.Right))
	}
	result = append(result, bottomLine)

	return strings.Join(result, "\n")
}

func (m Model) renderHelpOverlay() string {
	helpLines := []struct{ key, desc string }{
		{"", "Navigation"},
		{"1 / 2", "Focus list / detail panel"},
		{"j/k", "Navigate items"},
		{"Tab", "Switch tab"},
		{"< / >", "Resize panels"},
		{"Esc", "Back to list / clear filter"},
		{"", ""},
		{"", "Tasks (1: list, 2: detail)"},
		{"a / e / d", "Add / edit / delete"},
		{"s", "Cycle status / toggle subtask"},
		{"/", "Search / filter"},
		{"F1/F2/F3", "Sort date / due / priority"},
		{"F4", "Tag filter bar"},
		{"l", "Log time (detail)"},
		{"b", "View blockers (detail)"},
		{"", ""},
		{"", "Tools"},
		{"p", "Focus timer"},
		{"X", "Export"},
		{"G", "Statistics"},
		{"Ctrl+Z", "Undo"},
		{"?", "This help"},
		{"q", "Quit"},
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Keyboard Shortcuts"))
	lines = append(lines, "")
	for _, h := range helpLines {
		if h.key == "" && h.desc == "" {
			lines = append(lines, "")
			continue
		}
		if h.key == "" {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render(h.desc))
			continue
		}
		k := lipgloss.NewStyle().Foreground(cyan).Width(16).Render(h.key)
		d := lipgloss.NewStyle().Foreground(gray).Render(h.desc)
		lines = append(lines, k+d)
	}
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc or ? to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 3).
		Width(50).
		Render(content)
}
