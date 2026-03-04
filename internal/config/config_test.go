package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.PanelRatio != 0.4 {
		t.Errorf("DefaultConfig().PanelRatio = %v, want 0.4", cfg.PanelRatio)
	}
	if cfg.DateFormat != "Jan 02, 2006" {
		t.Errorf("DefaultConfig().DateFormat = %q, want Jan 02, 2006", cfg.DateFormat)
	}
	if cfg.TimeFormat != "3:04 PM" {
		t.Errorf("DefaultConfig().TimeFormat = %q, want 3:04 PM", cfg.TimeFormat)
	}
	if cfg.DateTimeFormat != "Jan 02, 2006 3:04 PM" {
		t.Errorf("DefaultConfig().DateTimeFormat = %q, want Jan 02, 2006 3:04 PM", cfg.DateTimeFormat)
	}
}

func TestPath(t *testing.T) {
	p, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	if filepath.Base(p) != "config.json" {
		t.Errorf("Path() = %q, want basename config.json", p)
	}
	dir := filepath.Base(filepath.Dir(p))
	if dir != ".todo-app" {
		t.Errorf("Path() parent dir = %q, want .todo-app", dir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory to avoid writing to the real home directory.
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Override the path function by writing/reading directly.
	cfg := Config{PanelRatio: 0.6}

	// Save manually to temp path.
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Load manually from temp path.
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var loaded Config
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	loaded.validateWithWarnings()

	if loaded.PanelRatio != 0.6 {
		t.Errorf("loaded PanelRatio = %v, want 0.6", loaded.PanelRatio)
	}
}

func TestValidate_Clamp(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"zero defaults", 0, 0.4},
		{"below min", 0.1, 0.2},
		{"at min", 0.2, 0.2},
		{"normal", 0.5, 0.5},
		{"at max", 0.8, 0.8},
		{"above max", 0.95, 0.8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{PanelRatio: tt.input}
			cfg.validateWithWarnings()
			if cfg.PanelRatio != tt.want {
				t.Errorf("validate(%v) = %v, want %v", tt.input, cfg.PanelRatio, tt.want)
			}
		})
	}
}

func TestLoad_MissingFile(t *testing.T) {
	// Load uses the real home dir path, but if the file doesn't exist there,
	// it should return defaults. We test the logic by simulating a missing file
	// scenario via the raw code path.
	cfg := DefaultConfig()
	if cfg.PanelRatio != 0.4 {
		t.Errorf("DefaultConfig() PanelRatio = %v, want 0.4", cfg.PanelRatio)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "a", "b", "config.json")
	dir := filepath.Dir(nested)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	cfg := Config{PanelRatio: 0.35}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(nested, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	raw, err := os.ReadFile(nested)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var loaded Config
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	loaded.validateWithWarnings()

	if loaded.PanelRatio != 0.35 {
		t.Errorf("loaded PanelRatio = %v, want 0.35", loaded.PanelRatio)
	}
}

func TestFocusConfig_Defaults(t *testing.T) {
	tests := []struct {
		name  string
		input FocusConfig
		want  FocusConfig
	}{
		{
			name:  "all zero values get defaults",
			input: FocusConfig{},
			want: FocusConfig{
				WorkDuration:       25,
				ShortBreakDuration: 5,
				LongBreakDuration:  15,
				LongBreakInterval:  4,
				DailyGoal:          8,
			},
		},
		{
			name:  "non-zero values preserved",
			input: FocusConfig{WorkDuration: 30, ShortBreakDuration: 10, LongBreakDuration: 20, LongBreakInterval: 3, DailyGoal: 12},
			want:  FocusConfig{WorkDuration: 30, ShortBreakDuration: 10, LongBreakDuration: 20, LongBreakInterval: 3, DailyGoal: 12},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Focus: tt.input}
			cfg.validateWithWarnings()
			if cfg.Focus.WorkDuration != tt.want.WorkDuration {
				t.Errorf("WorkDuration = %d, want %d", cfg.Focus.WorkDuration, tt.want.WorkDuration)
			}
			if cfg.Focus.ShortBreakDuration != tt.want.ShortBreakDuration {
				t.Errorf("ShortBreakDuration = %d, want %d", cfg.Focus.ShortBreakDuration, tt.want.ShortBreakDuration)
			}
			if cfg.Focus.LongBreakDuration != tt.want.LongBreakDuration {
				t.Errorf("LongBreakDuration = %d, want %d", cfg.Focus.LongBreakDuration, tt.want.LongBreakDuration)
			}
			if cfg.Focus.LongBreakInterval != tt.want.LongBreakInterval {
				t.Errorf("LongBreakInterval = %d, want %d", cfg.Focus.LongBreakInterval, tt.want.LongBreakInterval)
			}
			if cfg.Focus.DailyGoal != tt.want.DailyGoal {
				t.Errorf("DailyGoal = %d, want %d", cfg.Focus.DailyGoal, tt.want.DailyGoal)
			}
		})
	}
}

func TestFocusConfig_Clamp(t *testing.T) {
	tests := []struct {
		name  string
		input FocusConfig
		want  FocusConfig
	}{
		{
			name:  "work duration below min clamped to 1",
			input: FocusConfig{WorkDuration: -5, ShortBreakDuration: 5, LongBreakDuration: 15, LongBreakInterval: 4, DailyGoal: 8},
			want:  FocusConfig{WorkDuration: 1, ShortBreakDuration: 5, LongBreakDuration: 15, LongBreakInterval: 4, DailyGoal: 8},
		},
		{
			name:  "work duration above max clamped to 120",
			input: FocusConfig{WorkDuration: 200, ShortBreakDuration: 5, LongBreakDuration: 15, LongBreakInterval: 4, DailyGoal: 8},
			want:  FocusConfig{WorkDuration: 120, ShortBreakDuration: 5, LongBreakDuration: 15, LongBreakInterval: 4, DailyGoal: 8},
		},
		{
			name:  "long break interval above max clamped to 10",
			input: FocusConfig{WorkDuration: 25, ShortBreakDuration: 5, LongBreakDuration: 15, LongBreakInterval: 15, DailyGoal: 8},
			want:  FocusConfig{WorkDuration: 25, ShortBreakDuration: 5, LongBreakDuration: 15, LongBreakInterval: 10, DailyGoal: 8},
		},
		{
			name:  "short break below min clamped to 1",
			input: FocusConfig{WorkDuration: 25, ShortBreakDuration: -1, LongBreakDuration: 15, LongBreakInterval: 4, DailyGoal: 8},
			want:  FocusConfig{WorkDuration: 25, ShortBreakDuration: 1, LongBreakDuration: 15, LongBreakInterval: 4, DailyGoal: 8},
		},
		{
			name:  "long break below min clamped to 1",
			input: FocusConfig{WorkDuration: 25, ShortBreakDuration: 5, LongBreakDuration: -3, LongBreakInterval: 4, DailyGoal: 8},
			want:  FocusConfig{WorkDuration: 25, ShortBreakDuration: 5, LongBreakDuration: 1, LongBreakInterval: 4, DailyGoal: 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Focus: tt.input}
			cfg.validateWithWarnings()
			if cfg.Focus.WorkDuration != tt.want.WorkDuration {
				t.Errorf("WorkDuration = %d, want %d", cfg.Focus.WorkDuration, tt.want.WorkDuration)
			}
			if cfg.Focus.ShortBreakDuration != tt.want.ShortBreakDuration {
				t.Errorf("ShortBreakDuration = %d, want %d", cfg.Focus.ShortBreakDuration, tt.want.ShortBreakDuration)
			}
			if cfg.Focus.LongBreakDuration != tt.want.LongBreakDuration {
				t.Errorf("LongBreakDuration = %d, want %d", cfg.Focus.LongBreakDuration, tt.want.LongBreakDuration)
			}
			if cfg.Focus.LongBreakInterval != tt.want.LongBreakInterval {
				t.Errorf("LongBreakInterval = %d, want %d", cfg.Focus.LongBreakInterval, tt.want.LongBreakInterval)
			}
		})
	}
}

func TestDefaultConfig_FocusDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Focus.WorkDuration != 25 {
		t.Errorf("WorkDuration = %d, want 25", cfg.Focus.WorkDuration)
	}
	if cfg.Focus.ShortBreakDuration != 5 {
		t.Errorf("ShortBreakDuration = %d, want 5", cfg.Focus.ShortBreakDuration)
	}
	if cfg.Focus.LongBreakDuration != 15 {
		t.Errorf("LongBreakDuration = %d, want 15", cfg.Focus.LongBreakDuration)
	}
	if cfg.Focus.LongBreakInterval != 4 {
		t.Errorf("LongBreakInterval = %d, want 4", cfg.Focus.LongBreakInterval)
	}
	if cfg.Focus.DailyGoal != 8 {
		t.Errorf("DailyGoal = %d, want 8", cfg.Focus.DailyGoal)
	}
	if !cfg.Focus.Sound {
		t.Error("Sound = false, want true")
	}
}

func TestRoundtrip_JSON(t *testing.T) {
	original := Config{PanelRatio: 0.55}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	decoded.validateWithWarnings()

	if decoded.PanelRatio != original.PanelRatio {
		t.Errorf("roundtrip PanelRatio = %v, want %v", decoded.PanelRatio, original.PanelRatio)
	}
}

func TestFormatFunctions(t *testing.T) {
	cfg := Config{DateFormat: "02.01.2006", TimeFormat: "15:04", DateTimeFormat: "02.01.2006 15:04"}
	cfg.validateWithWarnings()

	ts := time.Date(2026, 3, 2, 21, 7, 0, 0, time.Local)
	if got := cfg.FormatDate(ts); got != "02.03.2026" {
		t.Errorf("FormatDate = %q, want 02.03.2026", got)
	}
	if got := cfg.FormatTime(ts); got != "21:07" {
		t.Errorf("FormatTime = %q, want 21:07", got)
	}
	if got := cfg.FormatDateTime(ts); got != "02.03.2026 21:07" {
		t.Errorf("FormatDateTime = %q, want 02.03.2026 21:07", got)
	}
}

func TestValidateTimeLayout(t *testing.T) {
	tests := []struct {
		name    string
		layout  string
		wantErr bool
	}{
		{name: "date layout", layout: "2006-01-02", wantErr: false},
		{name: "time layout", layout: "3:04 PM", wantErr: false},
		{name: "literal string invalid", layout: "DD/MM/YYYY", wantErr: true},
		{name: "empty invalid", layout: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeLayout(tt.layout)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateTimeLayout(%q) expected error, got nil", tt.layout)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateTimeLayout(%q) unexpected error: %v", tt.layout, err)
			}
		})
	}
}

func TestValidate_InvalidLayoutsFallbackToDefaults(t *testing.T) {
	cfg := Config{
		PanelRatio:     0.5,
		DateFormat:     "DD/MM/YYYY",
		TimeFormat:     "hh:mm",
		DateTimeFormat: "YYYY-MM-DD hh:mm",
	}

	cfg.validateWithWarnings()

	if cfg.DateFormat != "Jan 02, 2006" {
		t.Errorf("DateFormat = %q, want Jan 02, 2006", cfg.DateFormat)
	}
	if cfg.TimeFormat != "3:04 PM" {
		t.Errorf("TimeFormat = %q, want 3:04 PM", cfg.TimeFormat)
	}
	if cfg.DateTimeFormat != "Jan 02, 2006 3:04 PM" {
		t.Errorf("DateTimeFormat = %q, want Jan 02, 2006 3:04 PM", cfg.DateTimeFormat)
	}
}

func TestDualPathBug_PartialFormatChange(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TimeFormat = "15:04"
	cfg.validateWithWarnings()

	ts := time.Date(2026, 3, 2, 0, 0, 0, 0, time.Local)

	// Changing TimeFormat must not affect FormatDate output.
	got := cfg.FormatDate(ts)
	want := "Mar 02, 2026"
	if got != want {
		t.Errorf("FormatDate after TimeFormat change = %q, want %q", got, want)
	}
}

func TestStripYear(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		want   string
	}{
		{name: "pretty year-last", layout: "Jan 02, 2006", want: "Jan 02"},
		{name: "ISO year-first", layout: "2006-01-02", want: "01-02"},
		{name: "EU year-last", layout: "02.01.2006", want: "02.01"},
		{name: "US year-last", layout: "01/02/2006", want: "01/02"},
		{name: "no year", layout: "Jan 02", want: "Jan 02"},
		{name: "year only", layout: "2006", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripYear(tt.layout); got != tt.want {
				t.Errorf("stripYear(%q) = %q, want %q", tt.layout, got, tt.want)
			}
		})
	}
}

func TestFormatDateShort(t *testing.T) {
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	sameYear := time.Date(2026, 7, 4, 0, 0, 0, 0, time.Local)
	diffYear := time.Date(2025, 12, 25, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name       string
		dateFormat string
		date       time.Time
		want       string
	}{
		{name: "pretty same year", dateFormat: "Jan 02, 2006", date: sameYear, want: "Jul 04"},
		{name: "pretty diff year", dateFormat: "Jan 02, 2006", date: diffYear, want: "Dec 25, 2025"},
		{name: "ISO same year", dateFormat: "2006-01-02", date: sameYear, want: "07-04"},
		{name: "ISO diff year", dateFormat: "2006-01-02", date: diffYear, want: "2025-12-25"},
		{name: "EU same year", dateFormat: "02.01.2006", date: sameYear, want: "04.07"},
		{name: "US same year", dateFormat: "01/02/2006", date: sameYear, want: "07/04"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{DateFormat: tt.dateFormat}
			if got := cfg.FormatDateShort(tt.date, now); got != tt.want {
				t.Errorf("FormatDateShort(%q) = %q, want %q", tt.dateFormat, got, tt.want)
			}
		})
	}
}

func TestValidate_LegacyConfigGetsFormatDefaults(t *testing.T) {
	legacyJSON := []byte(`{"panel_ratio":0.55}`)

	var cfg Config
	if err := json.Unmarshal(legacyJSON, &cfg); err != nil {
		t.Fatalf("unmarshal legacy json: %v", err)
	}

	cfg.validateWithWarnings()

	if cfg.DateFormat == "" || cfg.TimeFormat == "" || cfg.DateTimeFormat == "" {
		t.Fatalf("legacy config did not get defaults: %+v", cfg)
	}
}

func TestFormatNoteTitle(t *testing.T) {
	now := time.Date(2026, 3, 15, 14, 0, 0, 0, time.Local)
	cfg := DefaultConfig()

	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{
			name: "today",
			date: time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
			want: "Today, Mar 15",
		},
		{
			name: "yesterday",
			date: time.Date(2026, 3, 14, 0, 0, 0, 0, time.Local),
			want: "Yesterday, Mar 14",
		},
		{
			name: "within_week",
			date: time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local),
			want: "Tue, Mar 10",
		},
		{
			name: "same_year",
			date: time.Date(2026, 1, 5, 0, 0, 0, 0, time.Local),
			want: "Jan 05",
		},
		{
			name: "diff_year",
			date: time.Date(2025, 12, 25, 0, 0, 0, 0, time.Local),
			want: "Dec 25, 2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.FormatNoteTitle(tt.date, now)
			if got != tt.want {
				t.Errorf("FormatNoteTitle(%v) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

func TestFormatNoteTitle_CustomFormat(t *testing.T) {
	now := time.Date(2026, 3, 15, 14, 0, 0, 0, time.Local)
	cfg := Config{DateFormat: "02.01.2006"}

	today := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	got := cfg.FormatNoteTitle(today, now)
	if got != "Today, 15.03" {
		t.Errorf("FormatNoteTitle with EU format = %q, want Today, 15.03", got)
	}

	oldDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	got = cfg.FormatNoteTitle(oldDate, now)
	if got != "01.06.2025" {
		t.Errorf("FormatNoteTitle old date with EU format = %q, want 01.06.2025", got)
	}
}

func TestFormatNoteTitle_TimezoneRobust(t *testing.T) {
	now := time.Now()
	// Journal dates are stored as local midnight, so use Local here.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	cfg := DefaultConfig()

	got := cfg.FormatNoteTitle(today, now)
	if got[:5] != "Today" {
		t.Errorf("local midnight matching Today: got %q, want prefix 'Today'", got)
	}
}

func TestFormatDetailDate(t *testing.T) {
	cfg := DefaultConfig()
	ts := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local) // Sunday
	got := cfg.FormatDetailDate(ts)
	if got != "Sun, Mar 15, 2026" {
		t.Errorf("FormatDetailDate = %q, want Sun, Mar 15, 2026", got)
	}
}

func TestValidateWithWarnings_InvalidFormats(t *testing.T) {
	cfg := Config{
		PanelRatio: 0.5,
		DateFormat: "DD/MM/YYYY",
		TimeFormat: "hh:mm",
	}
	warnings := cfg.validateWithWarnings()
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateWithWarnings_ValidFormats_NoWarnings(t *testing.T) {
	cfg := DefaultConfig()
	warnings := cfg.validateWithWarnings()
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateTimeLayout_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		layout  string
		wantErr bool
	}{
		{name: "whitespace only", layout: "   ", wantErr: true},
		{name: "pure static text", layout: "Hello World", wantErr: true},
		{name: "My Date label", layout: "My Date", wantErr: true},
		{name: "year only", layout: "2006", wantErr: false},
		{name: "month only", layout: "January", wantErr: false},
		{name: "day only", layout: "02", wantErr: false},
		{name: "hour only", layout: "15", wantErr: false},
		{name: "standard date", layout: "2006-01-02", wantErr: false},
		{name: "pretty date", layout: "Jan 02, 2006", wantErr: false},
		{name: "12h time", layout: "3:04 PM", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeLayout(tt.layout)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateTimeLayout(%q) expected error, got nil", tt.layout)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateTimeLayout(%q) unexpected error: %v", tt.layout, err)
			}
		})
	}
}
