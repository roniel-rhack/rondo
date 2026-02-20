package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Sparkline block characters from lowest to highest.
var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline renders a horizontal sparkline using unicode block characters.
// Data values are scaled to fit within the block character range. The output
// is truncated or padded to the given width. Uses Cyan color.
func RenderSparkline(data []int, width int) string {
	if len(data) == 0 || width <= 0 {
		return ""
	}

	// Resample data to fit width if needed.
	display := resample(data, width)

	maxVal := 0
	for _, v := range display {
		if v > maxVal {
			maxVal = v
		}
	}

	style := lipgloss.NewStyle().Foreground(Cyan)
	var sb strings.Builder
	for _, v := range display {
		idx := 0
		if maxVal > 0 {
			idx = v * (len(sparkBlocks) - 1) / maxVal
		}
		sb.WriteRune(sparkBlocks[idx])
	}

	return style.Render(sb.String())
}

// resample reduces or expands data to exactly n points using nearest-neighbor.
func resample(data []int, n int) []int {
	if n <= 0 {
		return nil
	}
	if len(data) <= n {
		return data
	}
	result := make([]int, n)
	for i := 0; i < n; i++ {
		srcIdx := i * len(data) / n
		if srcIdx >= len(data) {
			srcIdx = len(data) - 1
		}
		result[i] = data[srcIdx]
	}
	return result
}

// RenderPriorityBreakdown renders a horizontal stacked bar showing priority
// distribution. Colors match task priority: Green (Low), Yellow (Medium),
// Red (High), Magenta (Urgent). Includes a legend.
func RenderPriorityBreakdown(low, med, high, urgent int) string {
	total := low + med + high + urgent
	if total == 0 {
		return lipgloss.NewStyle().Foreground(Gray).Render("No tasks")
	}

	barWidth := 40
	segments := []struct {
		count int
		color lipgloss.Color
		label string
	}{
		{low, Green, "Low"},
		{med, Yellow, "Med"},
		{high, Red, "High"},
		{urgent, Magenta, "Urgent"},
	}

	var bar strings.Builder
	for _, seg := range segments {
		if seg.count == 0 {
			continue
		}
		w := seg.count * barWidth / total
		if w == 0 && seg.count > 0 {
			w = 1
		}
		bar.WriteString(lipgloss.NewStyle().Foreground(seg.color).Render(strings.Repeat("█", w)))
	}

	// Legend.
	var legend []string
	for _, seg := range segments {
		if seg.count > 0 {
			legend = append(legend,
				lipgloss.NewStyle().Foreground(seg.color).Render("█")+
					lipgloss.NewStyle().Foreground(Gray).Render(fmt.Sprintf(" %s:%d", seg.label, seg.count)))
		}
	}

	return bar.String() + "\n" + strings.Join(legend, "  ")
}

// RenderTagCloud renders tags with counts, sorted by frequency descending.
// Tag names are styled in Cyan, counts in Gray.
func RenderTagCloud(tags map[string]int) string {
	if len(tags) == 0 {
		return lipgloss.NewStyle().Foreground(Gray).Render("No tags")
	}

	type tagEntry struct {
		name  string
		count int
	}

	entries := make([]tagEntry, 0, len(tags))
	for name, count := range tags {
		entries = append(entries, tagEntry{name, count})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})

	tagStyle := lipgloss.NewStyle().Foreground(Cyan)
	countStyle := lipgloss.NewStyle().Foreground(Gray)

	var parts []string
	for _, e := range entries {
		parts = append(parts, tagStyle.Render(e.name)+countStyle.Render(fmt.Sprintf("(%d)", e.count)))
	}

	return strings.Join(parts, "  ")
}

// RenderJournalStreak renders a streak summary showing current streak,
// longest streak, and a sparkline of activity over the past N days.
func RenderJournalStreak(completionsByDay map[string]int, days int) string {
	if days <= 0 {
		days = 30
	}

	labelStyle := lipgloss.NewStyle().Foreground(Gray)
	valueStyle := lipgloss.NewStyle().Foreground(White).Bold(true)

	// Build ordered data for the past N days.
	today := time.Now().Truncate(24 * time.Hour)
	data := make([]int, days)
	for i := 0; i < days; i++ {
		day := today.AddDate(0, 0, -(days - 1 - i))
		key := day.Format(time.DateOnly)
		data[i] = completionsByDay[key]
	}

	// Calculate current streak (consecutive days with entries, ending today or yesterday).
	currentStreak := 0
	for i := days - 1; i >= 0; i-- {
		if data[i] > 0 {
			currentStreak++
		} else {
			// Allow today to be zero if yesterday had entries (streak still counts).
			if i == days-1 && i > 0 && data[i-1] > 0 {
				continue
			}
			break
		}
	}

	// Calculate longest streak.
	longestStreak := 0
	streak := 0
	for _, v := range data {
		if v > 0 {
			streak++
			if streak > longestStreak {
				longestStreak = streak
			}
		} else {
			streak = 0
		}
	}

	sparkline := RenderSparkline(data, days)

	return fmt.Sprintf("%s %s  %s %s\n%s",
		labelStyle.Render("Current streak:"),
		valueStyle.Render(fmt.Sprintf("%d days", currentStreak)),
		labelStyle.Render("Longest:"),
		valueStyle.Render(fmt.Sprintf("%d days", longestStreak)),
		sparkline,
	)
}
