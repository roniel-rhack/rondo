// internal/app/styles.go
package app

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/ui"
)

// Local aliases for convenience.
var (
	cyan  = ui.Cyan
	white = ui.White
	gray  = ui.Gray
)

var (
	// Dialog overlay
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(1, 2).
			Width(60)
)
