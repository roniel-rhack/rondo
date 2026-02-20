package task

import (
	"fmt"
	"time"
)

type TimeLog struct {
	ID       int64
	TaskID   int64
	Duration time.Duration
	Note     string
	LoggedAt time.Time
}

// ParseDuration parses a human-friendly duration string such as "1h30m",
// "45m", or "2h". It delegates to time.ParseDuration which already handles
// these formats. An error is returned if the string is empty or unparseable.
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", s, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("negative duration not allowed: %s", s)
	}
	return d, nil
}

// FormatDuration formats a duration as "Xh Ym". Zero-value components are
// omitted (e.g., 90 minutes becomes "1h 30m", 120 minutes becomes "2h",
// 45 minutes becomes "45m"). A zero duration returns "0m".
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	totalMinutes := int(d.Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

// TotalDuration sums the durations of all provided time logs.
func TotalDuration(logs []TimeLog) time.Duration {
	var total time.Duration
	for _, l := range logs {
		total += l.Duration
	}
	return total
}
