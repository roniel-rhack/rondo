package ui

import (
	"testing"
	"time"
)

func TestDueStatus_Overdue(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	if got := DueStatus(yesterday); got != DueOverdue {
		t.Errorf("DueStatus(yesterday) = %d, want DueOverdue (%d)", got, DueOverdue)
	}

	weekAgo := time.Now().AddDate(0, 0, -7)
	if got := DueStatus(weekAgo); got != DueOverdue {
		t.Errorf("DueStatus(weekAgo) = %d, want DueOverdue (%d)", got, DueOverdue)
	}
}

func TestDueStatus_Today(t *testing.T) {
	today := time.Now()
	if got := DueStatus(today); got != DueToday {
		t.Errorf("DueStatus(today) = %d, want DueToday (%d)", got, DueToday)
	}
}

func TestDueStatus_Soon(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	if got := DueStatus(tomorrow); got != DueSoon {
		t.Errorf("DueStatus(tomorrow) = %d, want DueSoon (%d)", got, DueSoon)
	}

	threeDays := time.Now().AddDate(0, 0, 3)
	if got := DueStatus(threeDays); got != DueSoon {
		t.Errorf("DueStatus(3 days) = %d, want DueSoon (%d)", got, DueSoon)
	}
}

func TestDueStatus_Far(t *testing.T) {
	nextWeek := time.Now().AddDate(0, 0, 7)
	if got := DueStatus(nextWeek); got != DueFar {
		t.Errorf("DueStatus(nextWeek) = %d, want DueFar (%d)", got, DueFar)
	}

	nextMonth := time.Now().AddDate(0, 1, 0)
	if got := DueStatus(nextMonth); got != DueFar {
		t.Errorf("DueStatus(nextMonth) = %d, want DueFar (%d)", got, DueFar)
	}
}

func TestDueStyle_ReturnsNonEmpty(t *testing.T) {
	levels := []DueLevel{DueNone, DueFar, DueSoon, DueToday, DueOverdue}
	for _, level := range levels {
		style := DueStyle(level)
		// Verify the style can render without panic.
		result := style.Render("test")
		if result == "" {
			t.Errorf("DueStyle(%d).Render(\"test\") returned empty string", level)
		}
	}
}

func TestDueBadge(t *testing.T) {
	tests := []struct {
		level DueLevel
		want  string
	}{
		{DueOverdue, "OVERDUE"},
		{DueToday, "TODAY"},
		{DueSoon, "SOON"},
		{DueFar, ""},
		{DueNone, ""},
	}
	for _, tt := range tests {
		got := DueBadge(tt.level)
		if got != tt.want {
			t.Errorf("DueBadge(%d) = %q, want %q", tt.level, got, tt.want)
		}
	}
}
