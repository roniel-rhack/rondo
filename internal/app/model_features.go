package app

import (
	"fmt"
	"os"
	"strings"
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
	switch m.focusPhase {
	case phaseWork, phaseBreak:
		m.mode = modeFocusConfirmCancel
		return m, nil
	case phaseWorkDone:
		return m.startBreak()
	case phaseBreakDone:
		return m.startWork()
	default:
		return m.startWork()
	}
}

func (m *Model) startWork() (tea.Model, tea.Cmd) {
	var taskID int64
	if sel := m.selectedTask(); sel != nil {
		taskID = sel.ID
	}
	session := &focus.Session{
		TaskID:    taskID,
		Kind:      focus.KindWork,
		Duration:  time.Duration(m.cfg.Focus.WorkDuration) * time.Minute,
		StartedAt: time.Now().UTC(),
		CyclePos:  m.focusCyclePos + 1,
	}
	if err := m.focusStore.Create(session); err != nil {
		return m, m.setError(err)
	}
	m.focusSession = session
	m.focusPhase = phaseWork
	m.mode = modeNormal
	return m, tea.Batch(
		m.setStatus("Focus session started"),
		focusTick(),
	)
}

func (m *Model) startBreak() (tea.Model, tea.Cmd) {
	kind := focus.KindShortBreak
	duration := time.Duration(m.cfg.Focus.ShortBreakDuration) * time.Minute
	if m.focusCyclePos == 0 {
		kind = focus.KindLongBreak
		duration = time.Duration(m.cfg.Focus.LongBreakDuration) * time.Minute
	}
	session := &focus.Session{
		Kind:      kind,
		Duration:  duration,
		StartedAt: time.Now().UTC(),
	}
	if err := m.focusStore.Create(session); err != nil {
		return m, m.setError(err)
	}
	m.focusSession = session
	m.focusPhase = phaseBreak
	m.mode = modeNormal
	return m, tea.Batch(
		m.setStatus("Break started"),
		focusTick(),
	)
}

func (m *Model) handlePhaseComplete() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.cfg.Focus.Sound {
		cmds = append(cmds, focusBell())
	}
	switch m.focusPhase {
	case phaseWork:
		m.focusCyclePos = (m.focusCyclePos + 1) % m.cfg.Focus.LongBreakInterval
		if m.cfg.Focus.AutoStartBreak {
			model, cmd := m.startBreak()
			cmds = append(cmds, cmd)
			return model, tea.Batch(cmds...)
		}
		m.focusPhase = phaseWorkDone
		m.mode = modeFocusSessionEnd
	case phaseBreak:
		m.focusPhase = phaseBreakDone
		m.mode = modeFocusBreakEnd
	}
	return m, tea.Batch(cmds...)
}

func focusBell() tea.Cmd {
	return func() tea.Msg {
		fmt.Fprint(os.Stderr, "\a")
		return nil
	}
}

func (m *Model) updateFocusSessionEnd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.startBreak()
	case "s", "esc":
		m.focusPhase = phaseIdle
		m.focusSession = nil
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) updateFocusBreakEnd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.startWork()
	case "s", "esc":
		m.focusPhase = phaseIdle
		m.focusSession = nil
		m.mode = modeNormal
	}
	return m, nil
}

func focusTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return focusTickMsg(t)
	})
}

func (m *Model) focusTimerStr() string {
	switch m.focusPhase {
	case phaseWork:
		if m.focusSession == nil {
			return ""
		}
		return "🍅 " + focus.FormatTimer(m.focusSession.Remaining(time.Now())) + " " + m.cycleIndicator()
	case phaseBreak:
		if m.focusSession == nil {
			return ""
		}
		remaining := m.focusSession.Remaining(time.Now())
		if m.focusSession.Kind == focus.KindLongBreak {
			return "🌿 " + focus.FormatTimer(remaining) + " " + m.cycleIndicator()
		}
		return "☕ " + focus.FormatTimer(remaining) + " " + m.cycleIndicator()
	case phaseWorkDone:
		return "🍅✓ Press p for break"
	case phaseBreakDone:
		return "☕✓ Press p to focus"
	}
	return ""
}

func (m *Model) cycleIndicator() string {
	interval := m.cfg.Focus.LongBreakInterval
	if interval <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < interval; i++ {
		if i < m.focusCyclePos {
			b.WriteRune('●')
		} else {
			b.WriteRune('○')
		}
	}
	return b.String()
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
