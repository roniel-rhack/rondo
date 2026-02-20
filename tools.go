//go:build tools

package tools

// This file pins dependencies that are used in the project but not yet
// imported in application code. It will be removed once real imports exist.

import (
	_ "github.com/charmbracelet/bubbles/list"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/huh"
	_ "github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
)
