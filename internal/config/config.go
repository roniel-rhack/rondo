package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultPanelRatio    = 0.4
	minPanelRatio        = 0.2
	maxPanelRatio        = 0.8
	defaultDateFormat    = "Jan 02, 2006"
	defaultTimeFormat    = "3:04 PM"
	layoutSentinelOneY   = 2009
	layoutSentinelTwoY   = 2021
	layoutSentinelOneMon = time.March
	layoutSentinelTwoMon = time.November
)

// DateFormatPresets maps friendly names to Go date layouts.
var DateFormatPresets = map[string]string{
	"iso":      "2006-01-02",
	"european": "02.01.2006",
	"eu":       "02.01.2006",
	"us":       "01/02/2006",
	"pretty":   "Jan 02, 2006",
}

// TimeFormatPresets maps friendly names to Go time layouts.
var TimeFormatPresets = map[string]string{
	"24h": "15:04",
	"12h": "3:04 PM",
}

// DateTimeFormatPresets maps friendly names to Go date+time layouts.
var DateTimeFormatPresets = map[string]string{
	"iso":      "2006-01-02 15:04",
	"european": "02.01.2006 15:04",
	"eu":       "02.01.2006 15:04",
	"us":       "01/02/2006 3:04 PM",
	"pretty":   "Jan 02, 2006 3:04 PM",
}

// ResolvePreset returns the layout for a preset name, or the value itself
// if it doesn't match any preset.
func ResolvePreset(val string, presets map[string]string) string {
	val = strings.TrimSpace(val)
	if preset, ok := presets[strings.ToLower(val)]; ok {
		return preset
	}
	return val
}

// FocusConfig holds pomodoro/focus timer settings.
type FocusConfig struct {
	WorkDuration       int  `json:"work_duration_min"`
	ShortBreakDuration int  `json:"short_break_duration_min"`
	LongBreakDuration  int  `json:"long_break_duration_min"`
	LongBreakInterval  int  `json:"long_break_interval"`
	DailyGoal          int  `json:"daily_goal"`
	AutoStartBreak     bool `json:"auto_start_break"`
	Sound              bool `json:"sound"`
}

// Config holds user-configurable settings for the application.
type Config struct {
	PanelRatio     float64     `json:"panel_ratio"`
	DateFormat     string      `json:"date_format"`
	TimeFormat     string      `json:"time_format"`
	DateTimeFormat string      `json:"datetime_format"`
	Focus          FocusConfig `json:"focus"`
}

// DefaultConfig returns a Config populated with default values.
func DefaultConfig() Config {
	return Config{
		PanelRatio:     defaultPanelRatio,
		DateFormat:     defaultDateFormat,
		TimeFormat:     defaultTimeFormat,
		DateTimeFormat: defaultDateFormat + " " + defaultTimeFormat,
		Focus: FocusConfig{
			WorkDuration:       25,
			ShortBreakDuration: 5,
			LongBreakDuration:  15,
			LongBreakInterval:  4,
			DailyGoal:          8,
			Sound:              true,
		},
	}
}

// validateWithWarnings clamps PanelRatio to the allowed range, applies
// defaults for zero values, and returns warnings for fields that were reset
// to defaults due to invalid values.
func (c *Config) validateWithWarnings() []string {
	var warnings []string

	if c.PanelRatio == 0 {
		c.PanelRatio = defaultPanelRatio
	}
	if c.PanelRatio < minPanelRatio {
		c.PanelRatio = minPanelRatio
	}
	if c.PanelRatio > maxPanelRatio {
		c.PanelRatio = maxPanelRatio
	}

	c.DateFormat = strings.TrimSpace(c.DateFormat)
	if c.DateFormat == "" {
		c.DateFormat = defaultDateFormat
	}
	if err := ValidateTimeLayout(c.DateFormat); err != nil {
		warnings = append(warnings, fmt.Sprintf("date_format %q is invalid, using default %q", c.DateFormat, defaultDateFormat))
		c.DateFormat = defaultDateFormat
	}

	c.TimeFormat = strings.TrimSpace(c.TimeFormat)
	if c.TimeFormat == "" {
		c.TimeFormat = defaultTimeFormat
	}
	if err := ValidateTimeLayout(c.TimeFormat); err != nil {
		warnings = append(warnings, fmt.Sprintf("time_format %q is invalid, using default %q", c.TimeFormat, defaultTimeFormat))
		c.TimeFormat = defaultTimeFormat
	}

	c.DateTimeFormat = strings.TrimSpace(c.DateTimeFormat)
	if c.DateTimeFormat == "" {
		c.DateTimeFormat = c.DateFormat + " " + c.TimeFormat
	}
	if err := ValidateTimeLayout(c.DateTimeFormat); err != nil {
		c.DateTimeFormat = c.DateFormat + " " + c.TimeFormat
	}

	if c.Focus.WorkDuration == 0 {
		c.Focus.WorkDuration = 25
	}
	if c.Focus.ShortBreakDuration == 0 {
		c.Focus.ShortBreakDuration = 5
	}
	if c.Focus.LongBreakDuration == 0 {
		c.Focus.LongBreakDuration = 15
	}
	if c.Focus.LongBreakInterval == 0 {
		c.Focus.LongBreakInterval = 4
	}
	if c.Focus.DailyGoal == 0 {
		c.Focus.DailyGoal = 8
	}

	// Clamp durations.
	if c.Focus.WorkDuration < 1 {
		c.Focus.WorkDuration = 1
	}
	if c.Focus.WorkDuration > 120 {
		c.Focus.WorkDuration = 120
	}
	if c.Focus.ShortBreakDuration < 1 {
		c.Focus.ShortBreakDuration = 1
	}
	if c.Focus.LongBreakDuration < 1 {
		c.Focus.LongBreakDuration = 1
	}
	if c.Focus.LongBreakInterval < 1 {
		c.Focus.LongBreakInterval = 1
	}
	if c.Focus.LongBreakInterval > 10 {
		c.Focus.LongBreakInterval = 10
	}

	return warnings
}

// ValidateTimeLayout validates that a Go time layout contains at least one
// actual time token and is not only static text (e.g. "DD/MM/YYYY").
func ValidateTimeLayout(layout string) error {
	layout = strings.TrimSpace(layout)
	if layout == "" {
		return fmt.Errorf("layout cannot be empty")
	}

	s1 := time.Date(layoutSentinelOneY, layoutSentinelOneMon, 17, 8, 23, 7, 0, time.UTC).Format(layout)
	s2 := time.Date(layoutSentinelTwoY, layoutSentinelTwoMon, 28, 20, 51, 43, 0, time.UTC).Format(layout)
	if s1 == layout && s2 == layout {
		return fmt.Errorf("not a valid Go time layout")
	}
	return nil
}

func (c Config) FormatDate(t time.Time) string {
	return t.Format(c.DateFormat)
}

func (c Config) FormatTime(t time.Time) string {
	return t.Format(c.TimeFormat)
}

func (c Config) FormatDateTime(t time.Time) string {
	return t.Format(c.DateTimeFormat)
}

// stripYear removes the year component ("2006") and its adjacent separator
// from a Go time layout string. It handles both year-first ("2006-01-02")
// and year-last ("Jan 02, 2006") patterns.
func stripYear(layout string) string {
	idx := strings.Index(layout, "2006")
	if idx < 0 {
		return layout
	}

	before := layout[:idx]
	after := layout[idx+4:]

	// Year-last: "Jan 02, 2006" → strip trailing separator before year
	if after == "" || !strings.ContainsRune(" ,.-/", rune(after[0])) {
		result := strings.TrimRight(before, " ,.-/")
		if result == "" {
			return strings.TrimLeft(after, " ,.-/")
		}
		return result + strings.TrimLeft(after, " ,.-/")
	}

	// Year-first: "2006-01-02" → strip leading separator after year
	return before + strings.TrimLeft(after, " ,.-/")
}

// FormatDateShort formats a date without the year when the date is in the
// same year as now, otherwise falls back to the full FormatDate.
func (c Config) FormatDateShort(t time.Time, now time.Time) string {
	if t.Year() == now.Year() {
		short := stripYear(c.DateFormat)
		if short == c.DateFormat {
			return c.FormatDate(t)
		}
		return t.Format(short)
	}
	return c.FormatDate(t)
}

// sameDay returns true if a and b fall on the same calendar day.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// FormatNoteTitle returns a human-friendly date label for a journal note.
// It uses relative labels ("Today", "Yesterday", weekday) for recent dates,
// and falls back to FormatDateShort or FormatDate for older ones.
func (c Config) FormatNoteTitle(date, now time.Time) string {
	yesterday := now.AddDate(0, 0, -1)
	weekAgo := now.AddDate(0, 0, -6)

	switch {
	case sameDay(date, now):
		return "Today, " + c.FormatDateShort(date, now)
	case sameDay(date, yesterday):
		return "Yesterday, " + c.FormatDateShort(date, now)
	case date.After(weekAgo):
		return fmt.Sprintf("%s, %s", date.Format("Mon"), c.FormatDateShort(date, now))
	case date.Year() == now.Year():
		return c.FormatDateShort(date, now)
	default:
		return c.FormatDate(date)
	}
}

// FormatDetailDate formats a date with the weekday prefix for use in detail
// panel titles (e.g. "Mon, Jan 02, 2006").
func (c Config) FormatDetailDate(t time.Time) string {
	return t.Format("Mon, " + c.DateFormat)
}

// Path returns the absolute path to the config file (~/.todo-app/config.json).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config path: %w", err)
	}
	return filepath.Join(home, ".todo-app", "config.json"), nil
}

// Load reads the configuration from the config file. If the file does not
// exist, it returns DefaultConfig without error.
func Load() (Config, error) {
	cfg, _, err := LoadWithWarnings()
	return cfg, err
}

// LoadWithWarnings is like Load but also returns warnings for invalid format
// fields that were reset to defaults.
func LoadWithWarnings() (Config, []string, error) {
	p, err := Path()
	if err != nil {
		return DefaultConfig(), nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil, nil
		}
		return DefaultConfig(), nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil, fmt.Errorf("parse config: %w", err)
	}
	warnings := cfg.validateWithWarnings()
	return cfg, warnings, nil
}

// Save writes the configuration to the config file as formatted JSON. It
// creates the parent directory if necessary.
func Save(cfg Config) error {
	_ = cfg.validateWithWarnings()

	p, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(p, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
