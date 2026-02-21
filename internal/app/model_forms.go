package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/roniel/todo-app/internal/task"
)

// updateFormMsg handles ALL message types for form modes (not just KeyMsg).
func (m *Model) updateFormMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize even during form mode
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.resizeComponents()
	}

	// Only intercept Esc from key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		m.formData = nil
		return m, nil
	}

	// Pass all messages to the form (cursor blink, timers, keys, etc.)
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			if m.mode == modeSubtask {
				selected := m.selectedTask()
				title := ""
				if m.formData != nil {
					title = m.formData.Title
				}
				if selected != nil && title != "" {
					if err := m.store.AddSubtask(selected.ID, title); err != nil {
						return m, m.setError(err)
					}
				}
				m.mode = modeNormal
				m.form = nil
				m.formData = nil
				if err := m.reload(); err != nil {
					return m, m.setError(err)
				}
				return m, m.setStatus("Subtask added")
			}
			if m.mode == modeEditSubtask {
				selected := m.selectedTask()
				title := ""
				if m.formData != nil {
					title = m.formData.Title
				}
				if selected != nil && title != "" && m.subtaskIdx >= 0 && m.subtaskIdx < len(selected.Subtasks) {
					st := selected.Subtasks[m.subtaskIdx]
					if err := m.store.UpdateSubtask(st.ID, title); err != nil {
						return m, m.setError(err)
					}
				}
				m.mode = modeNormal
				m.form = nil
				m.formData = nil
				if err := m.reload(); err != nil {
					return m, m.setError(err)
				}
				return m, m.setStatus("Subtask updated")
			}
			cmd := m.submitTaskForm()
			return m, cmd
		}
		if m.form.State == huh.StateAborted {
			m.mode = modeNormal
			m.form = nil
			m.formData = nil
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) updateExportForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.resizeComponents()
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		m.exportFormData = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			data := m.exportFormData
			m.mode = modeNormal
			m.form = nil
			m.exportFormData = nil
			return m, m.runExport(data)
		}
		if m.form.State == huh.StateAborted {
			m.mode = modeNormal
			m.form = nil
			m.exportFormData = nil
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) updateTimeLogForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.resizeComponents()
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		m.timeLogFormData = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			selected := m.selectedTask()
			if selected != nil && m.timeLogFormData != nil {
				dur, err := task.ParseDuration(m.timeLogFormData.Duration)
				if err == nil {
					if err := m.store.AddTimeLog(selected.ID, dur, m.timeLogFormData.Note); err != nil {
						m.mode = modeNormal
						m.form = nil
						m.timeLogFormData = nil
						return m, m.setError(err)
					}
				}
			}
			m.mode = modeNormal
			m.form = nil
			m.timeLogFormData = nil
			if err := m.reload(); err != nil {
				return m, m.setError(err)
			}
			return m, m.setStatus("Time logged")
		}
		if m.form.State == huh.StateAborted {
			m.mode = modeNormal
			m.form = nil
			m.timeLogFormData = nil
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		selected := m.selectedTask()
		if selected != nil {
			// Capture for undo.
			deletedTask := *selected
			m.undoAction = &undoAction{
				description: fmt.Sprintf("Undo delete %q", deletedTask.Title),
				undo: func() error {
					return m.store.Restore(&deletedTask)
				},
			}
			if err := m.store.Delete(selected.ID); err != nil {
				return m, m.setError(err)
			}
			if err := m.reload(); err != nil {
				m.mode = modeNormal
				return m, m.setError(err)
			}
		}
		m.mode = modeNormal
		return m, m.setStatus("Task deleted")
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) updateConfirmDeleteSubtask(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		selected := m.selectedTask()
		if selected != nil && m.subtaskIdx >= 0 && m.subtaskIdx < len(selected.Subtasks) {
			st := selected.Subtasks[m.subtaskIdx]
			// Capture for undo.
			taskID := selected.ID
			stTitle := st.Title
			stCompleted := st.Completed
			stPosition := st.Position
			m.undoAction = &undoAction{
				description: fmt.Sprintf("Undo delete subtask %q", stTitle),
				undo: func() error {
					return m.store.RestoreSubtask(taskID, stTitle, stCompleted, stPosition)
				},
			}
			if err := m.store.DeleteSubtask(st.ID); err != nil {
				m.mode = modeNormal
				return m, m.setError(err)
			}
			if m.subtaskIdx > 0 && m.subtaskIdx >= len(selected.Subtasks)-1 {
				m.subtaskIdx--
			}
			if err := m.reload(); err != nil {
				m.mode = modeNormal
				return m, m.setError(err)
			}
		}
		m.mode = modeNormal
		return m, m.setStatus("Subtask deleted")
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) updateFocusConfirmCancel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.focusActive = false
		m.focusSession = nil
		m.mode = modeNormal
		return m, m.setStatus("Focus session cancelled")
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) submitTaskForm() tea.Cmd {
	if m.formData == nil {
		return nil
	}

	var dueDate *time.Time
	if m.formData.DueDate != "" {
		if d, err := time.Parse(time.DateOnly, m.formData.DueDate); err == nil {
			dueDate = &d
		}
	}

	var tags []string
	if m.formData.Tags != "" {
		for _, t := range strings.Split(m.formData.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	recurFreq := task.ParseRecurFreq(m.formData.RecurFreq)

	var statusCmd tea.Cmd
	if m.mode == modeAdd {
		t := &task.Task{
			Title:         m.formData.Title,
			Description:   m.formData.Description,
			Priority:      m.formData.Priority,
			DueDate:       dueDate,
			Tags:          tags,
			RecurFreq:     recurFreq,
			RecurInterval: 1,
		}
		if err := m.store.Create(t); err != nil {
			statusCmd = m.setError(err)
		} else {
			if recurFreq != task.RecurNone {
				_ = m.store.UpdateRecurrence(t.ID, recurFreq, 1)
			}
			statusCmd = m.setStatus("Task created")
		}
	} else if m.mode == modeEdit {
		selected := m.selectedTask()
		if selected != nil {
			selected.Title = m.formData.Title
			selected.Description = m.formData.Description
			selected.Priority = m.formData.Priority
			selected.DueDate = dueDate
			selected.Tags = tags
			selected.RecurFreq = recurFreq
			if recurFreq != task.RecurNone && selected.RecurInterval == 0 {
				selected.RecurInterval = 1
			}
			if err := m.store.Update(selected); err != nil {
				statusCmd = m.setError(err)
			} else {
				_ = m.store.UpdateRecurrence(selected.ID, recurFreq, selected.RecurInterval)
				statusCmd = m.setStatus("Task updated")
			}
		}
	}

	m.mode = modeNormal
	m.form = nil
	m.formData = nil
	if err := m.reload(); err != nil {
		return m.setError(err)
	}
	return statusCmd
}
