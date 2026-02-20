package journal

import (
	"fmt"
	"time"
)

// Note represents a single day's journal. One note per calendar day.
type Note struct {
	ID        int64
	Date      time.Time // Truncated to date (year, month, day)
	Hidden    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Entries   []Entry
}

// DateTitle returns the human-readable date string used as the note title.
func (n Note) DateTitle() string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	yesterday := today.AddDate(0, 0, -1)
	weekAgo := today.AddDate(0, 0, -6)

	switch {
	case n.Date.Equal(today):
		return "Today, " + n.Date.Format("Jan 02")
	case n.Date.Equal(yesterday):
		return "Yesterday, " + n.Date.Format("Jan 02")
	case n.Date.After(weekAgo):
		return n.Date.Format("Mon, Jan 02")
	case n.Date.Year() == now.Year():
		return n.Date.Format("Jan 02")
	default:
		return n.Date.Format("Jan 02, 2006")
	}
}

// FilterValue implements list.Item for the bubbles list widget.
func (n Note) FilterValue() string {
	return n.DateTitle()
}

// Title implements list.DefaultItem.
func (n Note) Title() string {
	return n.DateTitle()
}

// Description implements list.DefaultItem.
func (n Note) Description() string {
	return fmt.Sprintf("%d entries", len(n.Entries))
}

// Entry is a single journal entry within a day's Note.
type Entry struct {
	ID        int64
	NoteID    int64
	Body      string
	CreatedAt time.Time
}
