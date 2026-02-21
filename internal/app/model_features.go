package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/roniel/todo-app/internal/focus"
)

func (m *Model) handleUndo() (tea.Model, tea.Cmd) {
	if m.undoAction == nil {
		return m, m.setStatus("Nothing to undo")
	}
	action := m.undoAction
	m.undoAction = nil
	if err := action.undo(); err != nil {
		return m, m.setError(err)
	}
	if err := m.reload(); err != nil {
		return m, m.setError(err)
	}
	if err := m.reloadJournal(); err != nil {
		return m, m.setError(err)
	}
	return m, m.setStatus("Undone: " + action.description)
}

func (m *Model) handleFocusToggle() (tea.Model, tea.Cmd) {
	if m.focusActive {
		m.mode = modeFocusConfirmCancel
		return m, nil
	}
	// Start a new focus session.
	var taskID int64
	if sel := m.selectedTask(); sel != nil {
		taskID = sel.ID
	}
	session := &focus.Session{
		TaskID:    taskID,
		Duration:  focus.DefaultDuration,
		StartedAt: time.Now(),
	}
	if err := m.focusStore.Create(session); err != nil {
		return m, m.setError(err)
	}
	m.focusSession = session
	m.focusActive = true
	return m, tea.Batch(
		m.setStatus("Focus session started (25 min)"),
		focusTick(),
	)
}

func focusTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return focusTickMsg(t)
	})
}

func (m *Model) focusTimerStr() string {
	if !m.focusActive || m.focusSession == nil {
		return ""
	}
	remaining := m.focusSession.Remaining(time.Now())
	return "🍅 " + focus.FormatTimer(remaining)
}

func (m *Model) updateTagFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tags := m.allTags()
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		return m, nil
	case "enter":
		m.mode = modeNormal
		m.refreshList()
		m.updateDetail()
		return m, nil
	case "j", "right", "l":
		// Cycle forward through tags.
		if m.activeTag == "" {
			if len(tags) > 0 {
				m.activeTag = tags[0]
			}
		} else {
			for i, t := range tags {
				if t == m.activeTag {
					if i+1 < len(tags) {
						m.activeTag = tags[i+1]
					} else {
						m.activeTag = "" // Wrap to "All"
					}
					break
				}
			}
		}
		m.refreshList()
		m.updateDetail()
		return m, nil
	case "k", "left", "h":
		// Cycle backward through tags.
		if m.activeTag == "" {
			if len(tags) > 0 {
				m.activeTag = tags[len(tags)-1]
			}
		} else {
			for i, t := range tags {
				if t == m.activeTag {
					if i-1 >= 0 {
						m.activeTag = tags[i-1]
					} else {
						m.activeTag = "" // Wrap to "All"
					}
					break
				}
			}
		}
		m.refreshList()
		m.updateDetail()
		return m, nil
	}
	return m, nil
}

func (m *Model) updateBlockerPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Simple blocker management: show info + toggle.
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		return m, nil
	}
	return m, nil
}
