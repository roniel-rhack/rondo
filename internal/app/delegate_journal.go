package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/config"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/ui"
)

// noteItem wraps a journal.Note with a pre-computed title used for both
// display rendering and list filtering.
type noteItem struct {
	journal.Note
	title string
}

func (n noteItem) FilterValue() string { return n.title }

type noteDelegate struct {
	cfg config.Config
}

func newNoteDelegate(cfg config.Config) noteDelegate {
	return noteDelegate{cfg: cfg}
}

func (d noteDelegate) Height() int  { return 1 }
func (d noteDelegate) Spacing() int { return 0 }

func (d noteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d noteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ni, ok := item.(noteItem)
	if !ok {
		return
	}
	n := ni.Note

	isSelected := index == m.Index()
	availWidth := m.Width()

	dateLabel := ni.title
	countLabel := fmt.Sprintf("%d entries", len(n.Entries))

	if n.Hidden {
		dimStyle := lipgloss.NewStyle().Foreground(ui.DimGray)
		prefix := "░ "
		if isSelected {
			prefix = "▸ "
		}
		left := dimStyle.Render(prefix + dateLabel)
		right := dimStyle.Render(countLabel)
		gap := availWidth - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
		line := left + strings.Repeat(" ", gap) + right
		if isSelected && ui.IsDark() {
			line = lipgloss.NewStyle().Background(ui.SelectionBg).Render(line)
		}
		fmt.Fprint(w, line)
		return
	}

	dateStyle := lipgloss.NewStyle().Foreground(ui.White)
	countStyle := lipgloss.NewStyle().Foreground(ui.Gray)

	var left string
	if isSelected {
		cursor := lipgloss.NewStyle().Foreground(ui.Cyan).Render("▸ ")
		left = cursor + dateStyle.Render(dateLabel)
	} else {
		left = "  " + dateStyle.Render(dateLabel)
	}

	right := countStyle.Render(countLabel)
	gap := availWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	if isSelected && ui.IsDark() {
		line = lipgloss.NewStyle().Background(ui.SelectionBg).Render(line)
	}
	fmt.Fprint(w, line)
}
