// internal/ui/views.go
package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/config"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
)

func labelStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Gray).Width(12) }
func valueStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(White) }
func titleStyle() lipgloss.Style { return lipgloss.NewStyle().Bold(true).Foreground(White) }

// RenderTabs renders the tab bar.
func RenderTabs(activeTab int, allCount, activeCount, doneCount, journalCount int, width int) string {
	tabs := []struct {
		label string
		count int
	}{
		{"All", allCount},
		{"Active", activeCount},
		{"Done", doneCount},
	}

	tabNormal := lipgloss.NewStyle().Padding(0, 2).Foreground(Gray)
	tabActive := lipgloss.NewStyle().Padding(0, 2).Foreground(Cyan).Bold(true).Reverse(true)

	var rendered []string
	appTitle := lipgloss.NewStyle().Bold(true).Foreground(Cyan).Padding(0, 1).Render("RonDO")
	rendered = append(rendered, appTitle)

	for i, t := range tabs {
		label := fmt.Sprintf("%s (%d)", t.label, t.count)
		if i == activeTab {
			rendered = append(rendered, tabActive.Render(label))
		} else {
			rendered = append(rendered, tabNormal.Render(label))
		}
	}

	// Divider between task tabs and journal tab.
	divider := lipgloss.NewStyle().Foreground(DimGray).Render(" │ ")
	rendered = append(rendered, divider)

	journalLabel := fmt.Sprintf("Journal (%d)", journalCount)
	if activeTab == 3 {
		rendered = append(rendered, tabActive.Render(journalLabel))
	} else {
		rendered = append(rendered, tabNormal.Render(journalLabel))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(DimGray).
		Width(width).
		Render(row)
}

// RenderDetail renders the task detail panel content.
// subtaskIdx indicates which subtask has the cursor (-1 for none).
// detailFocused controls whether the subtask cursor is shown.
// noteIdx and notesFocused control the notes cursor when in note navigation mode.
func RenderDetail(t *task.Task, width int, subtaskIdx int, detailFocused bool, noteIdx int, notesFocused bool, cfg config.Config) string {
	if t == nil {
		return lipgloss.NewStyle().
			Foreground(Gray).
			Align(lipgloss.Center).
			Width(width).
			Render("\n\n\nSelect a task to view details")
	}

	var sections []string

	// Title (wrap long titles to fill available width)
	titleWidth := max(width-lipgloss.Width(labelStyle().Render("")), 10)
	sections = append(sections, labelStyle().Render("Title")+titleStyle().Width(titleWidth).Render(t.Title))
	sections = append(sections, "")

	// Status
	statusStr := t.Status.Icon() + " " + t.Status.String()
	sections = append(sections, labelStyle().Render("Status")+statusStyle(t.Status).Render(statusStr))

	// Blocked badge
	if len(t.BlockedByIDs) > 0 {
		blocked := task.IsBlocked(t.BlockedByIDs, func(id int64) task.Status {
			// Default to Pending if we can't look up; the badge is informational.
			return task.Pending
		})
		if blocked {
			badge := lipgloss.NewStyle().Foreground(Red).Bold(true).Render(" [BLOCKED]")
			sections[len(sections)-1] += badge
		}
	}

	// Priority
	sections = append(sections, labelStyle().Render("Priority")+prioStyle(t.Priority).Render(t.Priority.String()))

	// Due date with overdue badge
	if t.DueDate != nil {
		level := DueStatus(*t.DueDate)
		dateStr := cfg.FormatDate(*t.DueDate)
		badge := DueBadge(level)
		if badge != "" {
			dateStr += " " + DueStyle(level).Render(badge)
		}
		sections = append(sections, labelStyle().Render("Due")+DueStyle(level).Render(dateStr))
	}

	// Recurrence
	if t.RecurFreq != task.RecurNone {
		recurStr := "↻ " + t.RecurFreq.String()
		if t.RecurInterval > 1 {
			recurStr += fmt.Sprintf(" (every %d)", t.RecurInterval)
		}
		sections = append(sections, labelStyle().Render("Recurrence")+valueStyle().Render(recurStr))
	}

	// Created
	created := cfg.FormatDate(t.CreatedAt)
	sections = append(sections, labelStyle().Render("Created")+valueStyle().Render(created))

	// Tags
	if len(t.Tags) > 0 {
		tagStr := strings.Join(t.Tags, ", ")
		sections = append(sections, labelStyle().Render("Tags")+valueStyle().Render(tagStr))
	}

	// Time logged
	if len(t.TimeLogs) > 0 {
		total := task.TotalDuration(t.TimeLogs)
		timeStr := task.FormatDuration(total)
		timeStr += fmt.Sprintf(" (%d entries)", len(t.TimeLogs))
		sections = append(sections, labelStyle().Render("Time Logged")+valueStyle().Render(timeStr))
	}

	// Description (rendered as markdown)
	if t.Description != "" {
		sections = append(sections, "")
		sections = append(sections, labelStyle().Render("Description"))
		sections = append(sections, RenderMarkdown(t.Description, width-2))
	}

	// Metadata
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
		sections = append(sections, labelStyle().Render("Metadata")+valueStyle().Render(strings.Join(pairs, ", ")))
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
		sections = append(sections, labelStyle().Render("Subtasks")+valueStyle().Render(fmt.Sprintf("%d/%d", doneCount, len(t.Subtasks))))
		sections = append(sections, renderProgressBar(doneCount, len(t.Subtasks), width-4))
		sections = append(sections, "")
		showSubCursor := detailFocused && !notesFocused
		for i, st := range t.Subtasks {
			prefix := "  "
			if showSubCursor && i == subtaskIdx {
				prefix = "▸ "
			}
			if st.Completed {
				sections = append(sections, lipgloss.NewStyle().Foreground(Green).Render(prefix+"[x] "+st.Title))
			} else {
				sections = append(sections, lipgloss.NewStyle().Foreground(White).Render(prefix+"[ ] "+st.Title))
			}
		}
	}

	// Notes
	if len(t.Notes) > 0 {
		sections = append(sections, "")
		sections = append(sections, labelStyle().Render("Notes")+valueStyle().Render(fmt.Sprintf("%d", len(t.Notes))))
		sections = append(sections, "")
		for i, n := range t.Notes {
			prefix := "  "
			if detailFocused && notesFocused && i == noteIdx {
				prefix = lipgloss.NewStyle().Foreground(Cyan).Render("▸ ")
			}
			ts := lipgloss.NewStyle().Foreground(Gray).Render(n.CreatedAt.Format("Jan 02 15:04"))
			sections = append(sections, prefix+ts+" "+n.Body)
		}
	}

	return strings.Join(sections, "\n")
}

// RenderJournalDetail renders the journal note detail panel content.
// entryIdx indicates which entry has the cursor. detailFocused controls whether the cursor is shown.
func RenderJournalDetail(note *journal.Note, width int, entryIdx int, detailFocused bool, cfg config.Config) string {
	if note == nil {
		return lipgloss.NewStyle().
			Foreground(Gray).
			Align(lipgloss.Center).
			Width(width).
			Render("\n\n\nSelect a note to view entries")
	}

	var sections []string

	// Date title.
	dateTitle := titleStyle().Render(cfg.FormatDetailDate(note.Date))
	if note.Hidden {
		badge := lipgloss.NewStyle().Foreground(Yellow).Render(" [hidden]")
		dateTitle += badge
	}
	sections = append(sections, dateTitle)
	sections = append(sections, "")

	if len(note.Entries) == 0 {
		sections = append(sections, lipgloss.NewStyle().Foreground(Gray).Render("No entries yet. Press 'a' to add one."))
	} else {
		separator := lipgloss.NewStyle().Foreground(DimGray).Render(
			strings.Repeat("─ ", width/4),
		)
		for i, entry := range note.Entries {
			prefix := "  "
			if detailFocused && i == entryIdx {
				prefix = lipgloss.NewStyle().Foreground(Cyan).Render("▸ ")
			}
			ts := cfg.FormatTime(entry.CreatedAt)
			timestamp := prefix + lipgloss.NewStyle().Bold(true).Foreground(Cyan).Render(ts)
			sections = append(sections, timestamp)
			sections = append(sections, "  "+RenderMarkdown(entry.Body, width-4))
			if i < len(note.Entries)-1 {
				sections = append(sections, "")
				sections = append(sections, separator)
				sections = append(sections, "")
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

	bar := lipgloss.NewStyle().Foreground(Cyan).Render(strings.Repeat("█", filled))
	bar += lipgloss.NewStyle().Foreground(DimGray).Render(strings.Repeat("░", empty))
	return "  " + bar
}

// RenderStatusBar renders the bottom status bar.
// focusedPanel: 0=list, 1=detail. Key hints adapt to the focused panel context.
// activeTab: 0-2 for task tabs, 3 for journal tab.
// timerStr: optional focus timer string (e.g. "🍅 12:34") shown when non-empty.
// undoAvailable: whether an undo action is available.
func RenderStatusBar(total, done, active int, width int, statusMsg string, focusedPanel int, activeTab int, timerStr string, undoAvailable bool, notesFocused bool) string {
	keyStyle := lipgloss.NewStyle().Foreground(Cyan)
	dimStyle := lipgloss.NewStyle().Foreground(Gray)
	panelStyle := lipgloss.NewStyle().Foreground(Cyan).Bold(true)

	// Journal tab (activeTab == 3).
	if activeTab == 3 {
		var left string
		if statusMsg != "" {
			color := Green
			if strings.HasPrefix(statusMsg, "Error:") {
				color = Red
			}
			left = lipgloss.NewStyle().Foreground(color).Bold(true).Render(" " + statusMsg)
		} else {
			left = dimStyle.Render(fmt.Sprintf(" %d notes | %d entries today", total, done))
		}

		// Focus timer.
		if timerStr != "" {
			left += "  " + lipgloss.NewStyle().Foreground(timerColor(timerStr)).Bold(true).Render(timerStr)
		}

		var panelLabel string
		if focusedPanel == 1 {
			panelLabel = panelStyle.Render("[2:Entries]")
		} else {
			panelLabel = panelStyle.Render("[1:Notes]")
		}

		var bindings []struct{ key, desc string }
		if focusedPanel == 1 {
			bindings = []struct{ key, desc string }{
				{"e", "edit"}, {"d", "del"}, {"a", "add"}, {"j/k", "nav"}, {"?", "help"},
			}
		} else {
			bindings = []struct{ key, desc string }{
				{"a", "add"}, {"h", "hide"}, {"H", "show hidden"}, {"/", "find"}, {"?", "help"},
			}
		}
		if undoAvailable {
			bindings = append(bindings, struct{ key, desc string }{"^Z", "undo"})
		}

		var parts []string
		parts = append(parts, panelLabel)
		for _, b := range bindings {
			parts = append(parts, keyStyle.Render(b.key)+dimStyle.Render(":"+b.desc))
		}
		right := strings.Join(parts, dimStyle.Render(" "))

		gap := width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			return left
		}
		return left + strings.Repeat(" ", gap) + right
	}

	// Task tabs (activeTab 0-2).
	var left string
	if statusMsg != "" {
		color := Green
		if strings.HasPrefix(statusMsg, "Error:") {
			color = Red
		}
		left = lipgloss.NewStyle().Foreground(color).Bold(true).Render(" " + statusMsg)
	} else {
		left = dimStyle.Render(fmt.Sprintf(" %d tasks | %d done | %d active", total, done, active))
	}

	// Focus timer.
	if timerStr != "" {
		left += "  " + lipgloss.NewStyle().Foreground(Red).Bold(true).Render(timerStr)
	}

	// Panel indicator.
	var panelLabel string
	if focusedPanel == 1 {
		panelLabel = panelStyle.Render("[2:Details]")
	} else {
		panelLabel = panelStyle.Render("[1:Tasks]")
	}

	// Context-sensitive key hints.
	var bindings []struct{ key, desc string }
	if focusedPanel == 1 && notesFocused {
		bindings = []struct{ key, desc string }{
			{"a", "add"}, {"e", "edit"}, {"d", "del"},
			{"n", "subtasks"}, {"j/k", "nav"}, {"?", "help"},
		}
	} else if focusedPanel == 1 {
		bindings = []struct{ key, desc string }{
			{"a", "add"}, {"e", "edit"}, {"d", "del"}, {"s", "toggle"},
			{"l", "log"}, {"b", "blockers"}, {"n", "notes"}, {"j/k", "nav"}, {"?", "help"},
		}
	} else {
		bindings = []struct{ key, desc string }{
			{"a", "add"}, {"e", "edit"}, {"d", "del"}, {"s", "status"},
			{"p", "focus"}, {"X", "export"}, {"/", "find"}, {"?", "help"},
		}
	}
	if undoAvailable {
		bindings = append(bindings, struct{ key, desc string }{"^Z", "undo"})
	}

	var parts []string
	parts = append(parts, panelLabel)
	for _, b := range bindings {
		parts = append(parts, keyStyle.Render(b.key)+dimStyle.Render(":"+b.desc))
	}
	right := strings.Join(parts, dimStyle.Render(" "))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

// RenderTagBar renders a horizontal tag filter bar.
func RenderTagBar(tags []string, activeTag string, width int) string {
	if len(tags) == 0 {
		return ""
	}
	tagNormal := lipgloss.NewStyle().Foreground(Gray).Padding(0, 1)
	tagActive := lipgloss.NewStyle().Foreground(Cyan).Bold(true).Reverse(true).Padding(0, 1)

	var rendered []string
	allLabel := "All"
	if activeTag == "" {
		rendered = append(rendered, tagActive.Render(allLabel))
	} else {
		rendered = append(rendered, tagNormal.Render(allLabel))
	}
	for _, tag := range tags {
		if tag == activeTag {
			rendered = append(rendered, tagActive.Render(tag))
		} else {
			rendered = append(rendered, tagNormal.Render(tag))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
	return lipgloss.NewStyle().Width(width).Render(row)
}

// RenderConfirmDialogBox renders a yes/no confirmation dialog box (without placement).
// An optional borderColor can be provided; defaults to Red.
func RenderConfirmDialogBox(title, message string, borderColor ...lipgloss.Color) string {
	color := Red
	if len(borderColor) > 0 {
		color = borderColor[0]
	}
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(White).
		Render(title) + "\n\n" +
		lipgloss.NewStyle().Foreground(Gray).Render(message) + "\n\n" +
		lipgloss.NewStyle().Foreground(Gray).Render("[y] confirm  [n/esc] cancel")

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(color).
		Padding(1, 2).
		Width(50).
		Render(content)
}

func statusStyle(s task.Status) lipgloss.Style {
	switch s {
	case task.InProgress:
		return lipgloss.NewStyle().Foreground(Yellow)
	case task.Done:
		return lipgloss.NewStyle().Foreground(Green)
	default:
		return lipgloss.NewStyle().Foreground(Gray)
	}
}

func prioStyle(p task.Priority) lipgloss.Style {
	switch p {
	case task.Urgent:
		return lipgloss.NewStyle().Foreground(Magenta)
	case task.High:
		return lipgloss.NewStyle().Foreground(Red)
	case task.Medium:
		return lipgloss.NewStyle().Foreground(Yellow)
	default:
		return lipgloss.NewStyle().Foreground(Green)
	}
}

// timerColor returns the display color for a focus timer string based on its emoji prefix.
func timerColor(timerStr string) lipgloss.Color {
	if strings.ContainsRune(timerStr, '🍅') {
		return Orange
	}
	if strings.ContainsRune(timerStr, '🌿') {
		return Cyan
	}
	if strings.ContainsRune(timerStr, '☕') {
		return Green
	}
	return Red
}

// RenderFocusProgressBar renders a horizontal progress bar for the focus timer.
// It shows elapsed/total as filled blocks with the given color.
func RenderFocusProgressBar(elapsed, total time.Duration, width int, color lipgloss.Color) string {
	if total <= 0 || width < 4 {
		return ""
	}
	barWidth := width - 2
	if barWidth > 40 {
		barWidth = 40
	}
	filled := 0
	if total > 0 {
		filled = int(int64(barWidth) * int64(elapsed) / int64(total))
	}
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled
	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled))
	bar += lipgloss.NewStyle().Foreground(DimGray).Render(strings.Repeat("░", empty))
	return "  " + bar
}
