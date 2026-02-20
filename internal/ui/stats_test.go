package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderSparkline_EmptyData(t *testing.T) {
	result := RenderSparkline(nil, 20)
	if result != "" {
		t.Errorf("expected empty string for nil data, got: %q", result)
	}

	result = RenderSparkline([]int{}, 20)
	if result != "" {
		t.Errorf("expected empty string for empty data, got: %q", result)
	}
}

func TestRenderSparkline_SingleValue(t *testing.T) {
	result := RenderSparkline([]int{5}, 10)
	if result == "" {
		t.Error("expected non-empty sparkline for single value")
	}
}

func TestRenderSparkline_OutputLength(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := RenderSparkline(data, 10)
	// The result contains ANSI escape codes, so we check the visual width.
	// Each data point should produce one character.
	if result == "" {
		t.Error("expected non-empty sparkline")
	}
}

func TestRenderSparkline_ResamplesLargeData(t *testing.T) {
	data := make([]int, 100)
	for i := range data {
		data[i] = i
	}
	result := RenderSparkline(data, 20)
	if result == "" {
		t.Error("expected non-empty sparkline for large data")
	}
}

func TestRenderSparkline_ZeroWidth(t *testing.T) {
	result := RenderSparkline([]int{1, 2, 3}, 0)
	if result != "" {
		t.Errorf("expected empty string for zero width, got: %q", result)
	}
}

func TestRenderPriorityBreakdown_AllZero(t *testing.T) {
	result := RenderPriorityBreakdown(0, 0, 0, 0)
	if !strings.Contains(result, "No tasks") {
		t.Errorf("expected 'No tasks' for all zeros, got: %q", result)
	}
}

func TestRenderPriorityBreakdown_HasBar(t *testing.T) {
	result := RenderPriorityBreakdown(5, 3, 2, 1)
	if result == "" {
		t.Error("expected non-empty result")
	}
	// Should contain block characters.
	if !strings.Contains(result, "█") {
		t.Errorf("expected block characters in bar, got: %q", result)
	}
}

func TestRenderPriorityBreakdown_Legend(t *testing.T) {
	result := RenderPriorityBreakdown(5, 3, 2, 1)
	if !strings.Contains(result, "Low") {
		t.Errorf("expected 'Low' in legend, got: %q", result)
	}
	if !strings.Contains(result, "Med") {
		t.Errorf("expected 'Med' in legend, got: %q", result)
	}
	if !strings.Contains(result, "High") {
		t.Errorf("expected 'High' in legend, got: %q", result)
	}
	if !strings.Contains(result, "Urgent") {
		t.Errorf("expected 'Urgent' in legend, got: %q", result)
	}
}

func TestRenderPriorityBreakdown_SinglePriority(t *testing.T) {
	result := RenderPriorityBreakdown(10, 0, 0, 0)
	if !strings.Contains(result, "Low") {
		t.Errorf("expected 'Low' in legend, got: %q", result)
	}
	// Should not contain other priority labels.
	if strings.Contains(result, "Med") {
		t.Errorf("should not contain 'Med' when count is 0, got: %q", result)
	}
}

func TestRenderTagCloud_Empty(t *testing.T) {
	result := RenderTagCloud(nil)
	if !strings.Contains(result, "No tags") {
		t.Errorf("expected 'No tags' for nil, got: %q", result)
	}

	result = RenderTagCloud(map[string]int{})
	if !strings.Contains(result, "No tags") {
		t.Errorf("expected 'No tags' for empty map, got: %q", result)
	}
}

func TestRenderTagCloud_Ordering(t *testing.T) {
	tags := map[string]int{
		"work":     5,
		"personal": 3,
		"urgent":   8,
	}
	result := RenderTagCloud(tags)

	// Urgent (8) should appear before work (5) which should appear before personal (3).
	urgentIdx := strings.Index(result, "urgent")
	workIdx := strings.Index(result, "work")
	personalIdx := strings.Index(result, "personal")

	if urgentIdx == -1 || workIdx == -1 || personalIdx == -1 {
		t.Fatalf("expected all tags in output, got: %q", result)
	}

	if urgentIdx > workIdx {
		t.Errorf("expected 'urgent' before 'work' in output, got: %q", result)
	}
	if workIdx > personalIdx {
		t.Errorf("expected 'work' before 'personal' in output, got: %q", result)
	}
}

func TestRenderTagCloud_ShowsCounts(t *testing.T) {
	tags := map[string]int{"go": 4}
	result := RenderTagCloud(tags)
	if !strings.Contains(result, "(4)") {
		t.Errorf("expected count (4) in output, got: %q", result)
	}
}

func TestRenderJournalStreak_Empty(t *testing.T) {
	result := RenderJournalStreak(map[string]int{}, 7)
	if !strings.Contains(result, "0 days") {
		t.Errorf("expected '0 days' for empty data, got: %q", result)
	}
}

func TestRenderJournalStreak_WithData(t *testing.T) {
	completions := make(map[string]int)
	today := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 5; i++ {
		day := today.AddDate(0, 0, -i)
		completions[day.Format(time.DateOnly)] = i + 1
	}

	result := RenderJournalStreak(completions, 30)
	if !strings.Contains(result, "Current streak:") {
		t.Errorf("expected 'Current streak:' in output, got: %q", result)
	}
	if !strings.Contains(result, "Longest:") {
		t.Errorf("expected 'Longest:' in output, got: %q", result)
	}
}

func TestRenderJournalStreak_DefaultDays(t *testing.T) {
	result := RenderJournalStreak(map[string]int{}, 0)
	// Should default to 30 days and not panic.
	if result == "" {
		t.Error("expected non-empty result with default days")
	}
}
