package task

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"1h30m", 90 * time.Minute, false},
		{"45m", 45 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"1h0m0s", time.Hour, false},
		{"30s", 30 * time.Second, false},
		{"", 0, true},
		{"abc", 0, true},
		{"-1h", 0, true},
	}
	for _, tt := range tests {
		got, err := ParseDuration(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseDuration(%q): expected error, got %v", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDuration(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		dur  time.Duration
		want string
	}{
		{90 * time.Minute, "1h 30m"},
		{2 * time.Hour, "2h"},
		{45 * time.Minute, "45m"},
		{0, "0m"},
		{time.Hour + 1*time.Minute, "1h 1m"},
		{3*time.Hour + 15*time.Minute, "3h 15m"},
		{-10 * time.Minute, "0m"},
	}
	for _, tt := range tests {
		if got := FormatDuration(tt.dur); got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.dur, got, tt.want)
		}
	}
}

func TestTotalDuration(t *testing.T) {
	logs := []TimeLog{
		{Duration: 30 * time.Minute},
		{Duration: 1 * time.Hour},
		{Duration: 15 * time.Minute},
	}
	got := TotalDuration(logs)
	want := time.Hour + 45*time.Minute
	if got != want {
		t.Errorf("TotalDuration = %v, want %v", got, want)
	}
}

func TestTotalDuration_Empty(t *testing.T) {
	got := TotalDuration(nil)
	if got != 0 {
		t.Errorf("TotalDuration(nil) = %v, want 0", got)
	}
}
