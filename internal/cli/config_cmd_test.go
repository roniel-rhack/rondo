package cli

import (
	"testing"

	"github.com/roniel/todo-app/internal/config"
)

func TestResolvePreset_DatePresets(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "iso", want: "2006-01-02"},
		{in: "EUROPEAN", want: "02.01.2006"},
		{in: "eu", want: "02.01.2006"},
		{in: "us", want: "01/02/2006"},
		{in: "02-01-2006", want: "02-01-2006"},
	}

	for _, tt := range tests {
		if got := config.ResolvePreset(tt.in, config.DateFormatPresets); got != tt.want {
			t.Fatalf("ResolvePreset(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestResolvePreset_TimePresets(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "24h", want: "15:04"},
		{in: "12h", want: "3:04 PM"},
		{in: "15:04:05", want: "15:04:05"},
	}

	for _, tt := range tests {
		if got := config.ResolvePreset(tt.in, config.TimeFormatPresets); got != tt.want {
			t.Fatalf("ResolvePreset(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestConfigDateSetter_AcceptsPreset(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := configKeys["date_format"].set(&cfg, "european"); err != nil {
		t.Fatalf("set date_format preset: %v", err)
	}
	if cfg.DateFormat != "02.01.2006" {
		t.Fatalf("DateFormat = %q, want 02.01.2006", cfg.DateFormat)
	}
}

func TestConfigTimeSetter_AcceptsPreset(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := configKeys["time_format"].set(&cfg, "12h"); err != nil {
		t.Fatalf("set time_format preset: %v", err)
	}
	if cfg.TimeFormat != "3:04 PM" {
		t.Fatalf("TimeFormat = %q, want 3:04 PM", cfg.TimeFormat)
	}
}

func TestConfigDateTimeSetter_AcceptsPreset(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := configKeys["datetime_format"].set(&cfg, "iso"); err != nil {
		t.Fatalf("set datetime_format preset: %v", err)
	}
	if cfg.DateTimeFormat != "2006-01-02 15:04" {
		t.Fatalf("DateTimeFormat = %q, want 2006-01-02 15:04", cfg.DateTimeFormat)
	}
}
