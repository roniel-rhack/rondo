// internal/app/styles.go
package app

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	cyan    = lipgloss.Color("#00BCD4")
	white   = lipgloss.Color("#FAFAFA")
	gray    = lipgloss.Color("#666666")
	dimGray = lipgloss.Color("#444444")
	black   = lipgloss.Color("#000000")
	red     = lipgloss.Color("#F44336")
	yellow  = lipgloss.Color("#FFC107")
	green   = lipgloss.Color("#4CAF50")
	magenta = lipgloss.Color("#E040FB")

	// App frame
	appStyle = lipgloss.NewStyle()

	// Header / tabs
	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(gray)

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(cyan).
			Bold(true).
			Reverse(true)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(dimGray)

	// Task list panel
	listPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			Padding(0, 1)

	listPanelFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(cyan).
				Padding(0, 1)

	// Detail panel
	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(gray).
				Padding(1, 2)

	// Detail content
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(white)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(gray).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(white)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(gray)

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(cyan)

	// Priority colors
	priorityLowStyle    = lipgloss.NewStyle().Foreground(green)
	priorityMedStyle    = lipgloss.NewStyle().Foreground(yellow)
	priorityHighStyle   = lipgloss.NewStyle().Foreground(red)
	priorityUrgentStyle = lipgloss.NewStyle().Foreground(magenta)

	// Status icons
	statusPendingStyle    = lipgloss.NewStyle().Foreground(gray)
	statusInProgressStyle = lipgloss.NewStyle().Foreground(yellow)
	statusDoneStyle       = lipgloss.NewStyle().Foreground(green)

	// Dialog overlay
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(1, 2).
			Width(60)

	// Subtask styles
	subtaskDoneStyle   = lipgloss.NewStyle().Foreground(green).Strikethrough(true)
	subtaskUndoneStyle = lipgloss.NewStyle().Foreground(white)

	// Help
	helpKeyStyle  = lipgloss.NewStyle().Foreground(cyan)
	helpDescStyle = lipgloss.NewStyle().Foreground(gray)
)
