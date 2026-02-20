package ui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DueLevel classifies how close a due date is.
type DueLevel int

const (
	DueNone    DueLevel = iota // No due date set
	DueFar                     // More than 3 days away
	DueSoon                    // Within 3 days
	DueToday                   // Due today
	DueOverdue                 // Past due
)

// DueStatus classifies a due date relative to today.
// DueNone should only be used by callers when there is no due date;
// this function always assumes a valid due date is provided.
func DueStatus(dueDate time.Time) DueLevel {
	today := time.Now().Truncate(24 * time.Hour)
	due := dueDate.Truncate(24 * time.Hour)

	diff := due.Sub(today)
	days := int(diff.Hours() / 24)

	switch {
	case days < 0:
		return DueOverdue
	case days == 0:
		return DueToday
	case days <= 3:
		return DueSoon
	default:
		return DueFar
	}
}

// DueStyle returns the appropriate lipgloss style for a due level.
func DueStyle(level DueLevel) lipgloss.Style {
	switch level {
	case DueOverdue:
		return lipgloss.NewStyle().Foreground(Red).Bold(true).Blink(true)
	case DueToday:
		return lipgloss.NewStyle().Foreground(Yellow).Bold(true)
	case DueSoon:
		return lipgloss.NewStyle().Foreground(Orange).Bold(true)
	case DueFar:
		return lipgloss.NewStyle().Foreground(Gray)
	default:
		return lipgloss.NewStyle().Foreground(Gray)
	}
}

// DueBadge returns a short badge string for the due level.
func DueBadge(level DueLevel) string {
	switch level {
	case DueOverdue:
		return "OVERDUE"
	case DueToday:
		return "TODAY"
	case DueSoon:
		return "SOON"
	default:
		return ""
	}
}
