package task

import (
	"testing"
	"time"
)

func TestRecurFreqString(t *testing.T) {
	tests := []struct {
		freq RecurFreq
		want string
	}{
		{RecurNone, "none"},
		{RecurDaily, "daily"},
		{RecurWeekly, "weekly"},
		{RecurMonthly, "monthly"},
		{RecurYearly, "yearly"},
		{RecurFreq(99), "none"},
	}
	for _, tt := range tests {
		if got := tt.freq.String(); got != tt.want {
			t.Errorf("RecurFreq(%d).String() = %q, want %q", tt.freq, got, tt.want)
		}
	}
}

func TestParseRecurFreq(t *testing.T) {
	tests := []struct {
		input string
		want  RecurFreq
	}{
		{"none", RecurNone},
		{"daily", RecurDaily},
		{"weekly", RecurWeekly},
		{"monthly", RecurMonthly},
		{"yearly", RecurYearly},
		{"", RecurNone},
		{"invalid", RecurNone},
	}
	for _, tt := range tests {
		if got := ParseRecurFreq(tt.input); got != tt.want {
			t.Errorf("ParseRecurFreq(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseRecurFreqRoundTrip(t *testing.T) {
	freqs := []RecurFreq{RecurNone, RecurDaily, RecurWeekly, RecurMonthly, RecurYearly}
	for _, f := range freqs {
		if got := ParseRecurFreq(f.String()); got != f {
			t.Errorf("round-trip failed for %d: String()=%q, ParseRecurFreq()=%d", f, f.String(), got)
		}
	}
}

func TestNextDueDate_Daily(t *testing.T) {
	base := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurDaily, RecurInterval: 1}
	got := NextDueDate(task)
	want := time.Date(2025, 3, 16, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate daily = %v, want %v", got, want)
	}
}

func TestNextDueDate_DailyInterval3(t *testing.T) {
	base := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurDaily, RecurInterval: 3}
	got := NextDueDate(task)
	want := time.Date(2025, 3, 18, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate daily*3 = %v, want %v", got, want)
	}
}

func TestNextDueDate_Weekly(t *testing.T) {
	base := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurWeekly, RecurInterval: 1}
	got := NextDueDate(task)
	want := time.Date(2025, 3, 22, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate weekly = %v, want %v", got, want)
	}
}

func TestNextDueDate_Monthly(t *testing.T) {
	base := time.Date(2025, 1, 31, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurMonthly, RecurInterval: 1}
	got := NextDueDate(task)
	// Jan 31 + 1 month = March 3 (Go's time.AddDate normalizes Feb overflow)
	want := time.Date(2025, 3, 3, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate monthly (jan31) = %v, want %v", got, want)
	}
}

func TestNextDueDate_MonthlyNormal(t *testing.T) {
	base := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurMonthly, RecurInterval: 2}
	got := NextDueDate(task)
	want := time.Date(2025, 5, 15, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate monthly*2 = %v, want %v", got, want)
	}
}

func TestNextDueDate_Yearly(t *testing.T) {
	base := time.Date(2024, 2, 29, 0, 0, 0, 0, time.Local) // leap day
	task := Task{DueDate: &base, RecurFreq: RecurYearly, RecurInterval: 1}
	got := NextDueDate(task)
	// Feb 29 2024 + 1 year = March 1 2025 (not a leap year)
	want := time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate yearly (leap) = %v, want %v", got, want)
	}
}

func TestNextDueDate_NoDueDate(t *testing.T) {
	task := Task{RecurFreq: RecurDaily, RecurInterval: 1}
	before := time.Now()
	got := NextDueDate(task)
	after := time.Now().Add(24 * time.Hour).Add(time.Second)
	if got.Before(before) || got.After(after) {
		t.Errorf("NextDueDate with no DueDate: got %v, expected near %v + 1 day", got, before)
	}
}

func TestNextDueDate_None(t *testing.T) {
	base := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurNone}
	got := NextDueDate(task)
	if !got.Equal(base) {
		t.Errorf("NextDueDate RecurNone = %v, want %v (unchanged)", got, base)
	}
}

func TestNextDueDate_ZeroInterval(t *testing.T) {
	base := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	task := Task{DueDate: &base, RecurFreq: RecurDaily, RecurInterval: 0}
	got := NextDueDate(task)
	want := time.Date(2025, 6, 2, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("NextDueDate zero interval = %v, want %v", got, want)
	}
}
