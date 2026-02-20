// internal/ui/views.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/task"
)

var (
	cyan    = lipgloss.Color("#00BCD4")
	white   = lipgloss.Color("#FAFAFA")
	gray    = lipgloss.Color("#666666")
	dimGray = lipgloss.Color("#444444")
	green   = lipgloss.Color("#4CAF50")
	red     = lipgloss.Color("#F44336")
	yellow  = lipgloss.Color("#FFC107")
	magenta = lipgloss.Color("#E040FB")

	labelStyle = lipgloss.NewStyle().Foreground(gray).Width(12)
	valueStyle = lipgloss.NewStyle().Foreground(white)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(white)
)

// RenderTabs renders the tab bar.
func RenderTabs(activeTab int, allCount, activeCount, doneCount int, width int) string {
	tabs := []struct {
		label string
		count int
	}{
		{"All", allCount},
		{"Active", activeCount},
		{"Done", doneCount},
	}

	tabNormal := lipgloss.NewStyle().Padding(0, 2).Foreground(gray)
	tabActive := lipgloss.NewStyle().Padding(0, 2).Foreground(cyan).Bold(true).Reverse(true)

	var rendered []string
	appTitle := lipgloss.NewStyle().Bold(true).Foreground(cyan).Padding(0, 1).Render("Todo")
	rendered = append(rendered, appTitle)

	for i, t := range tabs {
		label := fmt.Sprintf("%s (%d)", t.label, t.count)
		if i == activeTab {
			rendered = append(rendered, tabActive.Render(label))
		} else {
			rendered = append(rendered, tabNormal.Render(label))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(dimGray).
		Width(width).
		Render(row)
}

// RenderDetail renders the task detail panel content.
func RenderDetail(t *task.Task, width int) string {
	if t == nil {
		return lipgloss.NewStyle().
			Foreground(gray).
			Align(lipgloss.Center).
			Width(width).
			Render("\n\n\nSelect a task to view details")
	}

	var sections []string

	// Title
	sections = append(sections, titleStyle.Render(t.Title))
	sections = append(sections, "")

	// Status
	statusStr := t.Status.Icon() + " " + t.Status.String()
	sections = append(sections, labelStyle.Render("Status")+statusStyle(t.Status).Render(statusStr))

	// Priority
	sections = append(sections, labelStyle.Render("Priority")+prioStyle(t.Priority).Render(t.Priority.String()))

	// Due date
	if t.DueDate != nil {
		sections = append(sections, labelStyle.Render("Due")+valueStyle.Render(t.DueDate.Format("Jan 02, 2006")))
	}

	// Created
	sections = append(sections, labelStyle.Render("Created")+valueStyle.Render(t.CreatedAt.Format("Jan 02, 2006")))

	// Tags
	if len(t.Tags) > 0 {
		tagStr := strings.Join(t.Tags, ", ")
		sections = append(sections, labelStyle.Render("Tags")+valueStyle.Render(tagStr))
	}

	// Description
	if t.Description != "" {
		sections = append(sections, "")
		sections = append(sections, labelStyle.Render("Description"))
		sections = append(sections, valueStyle.Render(t.Description))
	}

	// Subtasks
	if len(t.Subtasks) > 0 {
		sections = append(sections, "")
		doneCount := 0
		for _, st := range t.Subtasks {
			if st.Completed {
				doneCount++
			}
		}
		sections = append(sections, labelStyle.Render("Subtasks")+valueStyle.Render(fmt.Sprintf("%d/%d", doneCount, len(t.Subtasks))))
		sections = append(sections, renderProgressBar(doneCount, len(t.Subtasks), width-4))
		sections = append(sections, "")
		for _, st := range t.Subtasks {
			if st.Completed {
				sections = append(sections, lipgloss.NewStyle().Foreground(green).Render("  [x] "+st.Title))
			} else {
				sections = append(sections, lipgloss.NewStyle().Foreground(white).Render("  [ ] "+st.Title))
			}
		}
	}

	return strings.Join(sections, "\n")
}

func renderProgressBar(done, total, width int) string {
	if total == 0 || width < 4 {
		return ""
	}
	barWidth := width - 2
	if barWidth > 40 {
		barWidth = 40
	}
	filled := barWidth * done / total
	empty := barWidth - filled

	bar := lipgloss.NewStyle().Foreground(cyan).Render(strings.Repeat("█", filled))
	bar += lipgloss.NewStyle().Foreground(dimGray).Render(strings.Repeat("░", empty))
	return "  " + bar
}

// RenderStatusBar renders the bottom status bar.
func RenderStatusBar(total, done, inProgress int, width int) string {
	keyStyle := lipgloss.NewStyle().Foreground(cyan)
	dimStyle := lipgloss.NewStyle().Foreground(gray)

	left := dimStyle.Render(fmt.Sprintf(" %d tasks | %d done | %d active", total, done, inProgress))

	bindings := []struct{ key, desc string }{
		{"a", "add"}, {"e", "edit"}, {"d", "del"}, {"s", "status"},
		{"t", "sub"}, {"/", "find"}, {"?", "help"},
	}
	var parts []string
	for _, b := range bindings {
		parts = append(parts, keyStyle.Render(b.key)+dimStyle.Render(":"+b.desc))
	}
	right := strings.Join(parts, dimStyle.Render(" "))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// RenderConfirmDialog renders a yes/no confirmation dialog.
func RenderConfirmDialog(title, message string, width, height int) string {
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(white).
		Render(title) + "\n\n" +
		lipgloss.NewStyle().Foreground(gray).Render(message) + "\n\n" +
		lipgloss.NewStyle().Foreground(gray).Render("[y] confirm  [n/esc] cancel")

	dialog := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(red).
		Padding(1, 2).
		Width(50).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}

func statusStyle(s task.Status) lipgloss.Style {
	switch s {
	case task.InProgress:
		return lipgloss.NewStyle().Foreground(yellow)
	case task.Done:
		return lipgloss.NewStyle().Foreground(green)
	default:
		return lipgloss.NewStyle().Foreground(gray)
	}
}

func prioStyle(p task.Priority) lipgloss.Style {
	switch p {
	case task.Urgent:
		return lipgloss.NewStyle().Foreground(magenta)
	case task.High:
		return lipgloss.NewStyle().Foreground(red)
	case task.Medium:
		return lipgloss.NewStyle().Foreground(yellow)
	default:
		return lipgloss.NewStyle().Foreground(green)
	}
}
