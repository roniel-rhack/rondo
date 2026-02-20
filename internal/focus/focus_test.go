package focus

import (
	"testing"
	"time"
)

func TestIsCompleted(t *testing.T) {
	s := Session{}
	if s.IsCompleted() {
		t.Error("expected incomplete session when CompletedAt is nil")
	}

	now := time.Now()
	s.CompletedAt = &now
	if !s.IsCompleted() {
		t.Error("expected completed session when CompletedAt is set")
	}
}

func TestElapsed(t *testing.T) {
	start := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s := Session{
		Duration:  25 * time.Minute,
		StartedAt: start,
	}

	tests := []struct {
		name string
		now  time.Time
		want time.Duration
	}{
		{"at start", start, 0},
		{"5 min in", start.Add(5 * time.Minute), 5 * time.Minute},
		{"25 min (exact)", start.Add(25 * time.Minute), 25 * time.Minute},
		{"30 min (capped)", start.Add(30 * time.Minute), 25 * time.Minute},
		{"before start", start.Add(-1 * time.Minute), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Elapsed(tt.now)
			if got != tt.want {
				t.Errorf("Elapsed(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestRemaining(t *testing.T) {
	start := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s := Session{
		Duration:  25 * time.Minute,
		StartedAt: start,
	}

	tests := []struct {
		name string
		now  time.Time
		want time.Duration
	}{
		{"at start", start, 25 * time.Minute},
		{"10 min in", start.Add(10 * time.Minute), 15 * time.Minute},
		{"25 min (exact)", start.Add(25 * time.Minute), 0},
		{"30 min (past)", start.Add(30 * time.Minute), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Remaining(tt.now)
			if got != tt.want {
				t.Errorf("Remaining(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestFormatTimer(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"25 minutes", 25 * time.Minute, "25:00"},
		{"4 min 30 sec", 4*time.Minute + 30*time.Second, "04:30"},
		{"zero", 0, "00:00"},
		{"59 seconds", 59 * time.Second, "00:59"},
		{"negative", -5 * time.Minute, "00:00"},
		{"1 hour", 60 * time.Minute, "60:00"},
		{"90 seconds", 90 * time.Second, "01:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimer(tt.d)
			if got != tt.want {
				t.Errorf("FormatTimer(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
