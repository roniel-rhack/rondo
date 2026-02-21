package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/roniel/todo-app/internal/export"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
)

func (m *Model) selectedTask() *task.Task {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	t, ok := item.(task.Task)
	if !ok {
		return nil
	}
	// Find the original task in our slice so that mutations persist.
	for i := range m.tasks {
		if m.tasks[i].ID == t.ID {
			return &m.tasks[i]
		}
	}
	return nil
}

func (m *Model) reload() error {
	var selectedID int64
	if sel := m.selectedTask(); sel != nil {
		selectedID = sel.ID
	}

	tasks, err := m.store.List()
	if err != nil {
		return err
	}
	m.tasks = tasks
	m.sortTasks()

	if selectedID != 0 {
		for i, item := range m.list.Items() {
			if t, ok := item.(task.Task); ok && t.ID == selectedID {
				m.list.Select(i)
				break
			}
		}
	}
	m.updateDetail()
	return nil
}

func (m *Model) refreshList() {
	filtered := m.filteredTasks()
	items := make([]list.Item, len(filtered))
	for i, t := range filtered {
		items[i] = t
	}
	m.list.SetItems(items)
}

func (m *Model) filteredTasks() []task.Task {
	var result []task.Task

	// First, filter by tab.
	for _, t := range m.tasks {
		switch m.activeTab {
		case 0: // All
			result = append(result, t)
		case 1: // Active (Pending + InProgress)
			if t.Status != task.Done {
				result = append(result, t)
			}
		case 2: // Done
			if t.Status == task.Done {
				result = append(result, t)
			}
		}
	}

	// Then, filter by active tag if set.
	if m.activeTag != "" {
		var tagFiltered []task.Task
		for _, t := range result {
			for _, tag := range t.Tags {
				if tag == m.activeTag {
					tagFiltered = append(tagFiltered, t)
					break
				}
			}
		}
		result = tagFiltered
	}

	return result
}

func (m *Model) updateDetail() {
	selected := m.selectedTask()
	// Clamp subtaskIdx to valid range after task changes, tab switches, or reloads.
	if selected == nil || len(selected.Subtasks) == 0 {
		m.subtaskIdx = 0
	} else if m.subtaskIdx >= len(selected.Subtasks) {
		m.subtaskIdx = len(selected.Subtasks) - 1
	}
	content := ui.RenderDetail(selected, m.viewport.Width, m.subtaskIdx, m.focusedPanel == 1)
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m *Model) sortTasks() {
	switch m.sortBy {
	case sortDue:
		sort.Slice(m.tasks, func(i, j int) bool {
			if m.tasks[i].DueDate == nil && m.tasks[j].DueDate == nil {
				return m.tasks[i].CreatedAt.After(m.tasks[j].CreatedAt)
			}
			if m.tasks[i].DueDate == nil {
				return false
			}
			if m.tasks[j].DueDate == nil {
				return true
			}
			return m.tasks[i].DueDate.Before(*m.tasks[j].DueDate)
		})
	case sortPriority:
		sort.Slice(m.tasks, func(i, j int) bool {
			return m.tasks[i].Priority > m.tasks[j].Priority
		})
	default: // sortCreated
		sort.Slice(m.tasks, func(i, j int) bool {
			return m.tasks[i].CreatedAt.After(m.tasks[j].CreatedAt)
		})
	}
	m.refreshList()
	m.updateDetail()
}

func (m *Model) allTags() []string {
	tagSet := make(map[string]bool)
	for _, t := range m.tasks {
		for _, tag := range t.Tags {
			tagSet[tag] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func (m *Model) runExport(data *ui.ExportFormData) tea.Cmd {
	return func() tea.Msg {
		home, err := os.UserHomeDir()
		if err != nil {
			return exportDoneMsg{err: err}
		}
		dir := filepath.Join(home, ".todo-app", "exports")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return exportDoneMsg{err: err}
		}

		ext := "md"
		if data.Format == "json" {
			ext = "json"
		}
		filename := fmt.Sprintf("export-%s.%s", time.Now().Format("2006-01-02"), ext)
		path := filepath.Join(dir, filename)

		f, err := os.Create(path)
		if err != nil {
			return exportDoneMsg{err: err}
		}
		defer f.Close()

		tasks, err := m.store.List()
		if err != nil {
			return exportDoneMsg{err: err}
		}

		if data.Format == "json" {
			var notes []journal.Note
			if data.IncludeJournal {
				notes, _ = m.journalStore.ListNotes(true)
			}
			err = export.WriteJSON(f, tasks, notes)
		} else {
			err = export.WriteTasks(f, tasks)
			if err == nil && data.IncludeJournal {
				notes, _ := m.journalStore.ListNotes(true)
				_, err = f.WriteString("\n---\n\n")
				if err == nil {
					err = export.WriteNotes(f, notes)
				}
			}
		}
		if err != nil {
			return exportDoneMsg{err: err}
		}
		return exportDoneMsg{path: path}
	}
}
