package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/config"
	"github.com/roniel/todo-app/internal/export"
	"github.com/roniel/todo-app/internal/focus"
	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
)

type mode int

const (
	modeNormal mode = iota
	modeAdd
	modeEdit
	modeConfirmDelete
	modeConfirmDeleteSubtask
	modeSubtask
	modeEditSubtask
	modeHelp
	modeJournalAdd
	modeJournalEdit
	modeJournalConfirmHide
	modeJournalConfirmDelete
	modeExport
	modeTimeLog
	modeFocusConfirmCancel
	modeTagFilter
	modeStats
	modeBlockerPicker
)

const tabCount = 4

type sortOrder int

const (
	sortCreated sortOrder = iota
	sortDue
	sortPriority
)

// undoAction stores a reversible action.
type undoAction struct {
	description string
	undo        func() error
}

// statsData holds computed statistics for the stats overlay.
type statsData struct {
	totalTasks    int
	doneTasks     int
	activeTasks   int
	lowCount      int
	medCount      int
	highCount     int
	urgentCount   int
	tagCounts     map[string]int
	focusToday    int
	journalStreak string
}

// Model is the top-level Bubbletea model for the todo application.
type Model struct {
	store    *task.Store
	tasks    []task.Task
	list     list.Model
	viewport viewport.Model
	help     help.Model
	form     *huh.Form
	formData *ui.TaskFormData

	// Journal fields.
	journalStore    *journal.Store
	notes           []journal.Note
	journalList     list.Model
	journalViewport viewport.Model
	journalFormData *ui.JournalFormData
	showHidden      bool
	entryIdx        int

	// Config + panel resize.
	cfg        config.Config
	panelRatio float64

	// Export.
	exportFormData *ui.ExportFormData

	// Time log.
	timeLogFormData *ui.TimeLogFormData

	// Focus timer.
	focusStore   *focus.Store
	focusSession *focus.Session
	focusActive  bool

	// Tag filter.
	activeTag     string
	tagBarVisible bool

	// Undo.
	undoAction *undoAction

	// Stats.
	stats *statsData

	mode         mode
	activeTab    int // 0=All, 1=Active, 2=Done, 3=Journal
	focusedPanel int // 0=list, 1=detail
	subtaskIdx   int
	sortBy       sortOrder
	width        int
	height       int
	ready        bool
	statusMsg    string
}

// tasksLoaded is a message sent after the initial data load completes.
type tasksLoaded struct {
	tasks []task.Task
	err   error
}

type clearStatusMsg struct{}

type focusTickMsg time.Time

type exportDoneMsg struct {
	path string
	err  error
}

func (m *Model) setStatus(msg string) tea.Cmd {
	m.statusMsg = msg
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m *Model) setError(err error) tea.Cmd {
	m.statusMsg = "Error: " + err.Error()
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// New creates a new Model backed by the given stores.
func New(store *task.Store, journalStore *journal.Store, focusStore *focus.Store, cfg config.Config) Model {
	delegate := newTaskDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(false)
	l.DisableQuitKeybindings()

	jDelegate := newNoteDelegate()
	jl := list.New(nil, jDelegate, 0, 0)
	jl.SetShowTitle(false)
	jl.SetShowHelp(false)
	jl.SetShowStatusBar(false)
	jl.SetFilteringEnabled(true)
	jl.SetShowFilter(false)
	jl.DisableQuitKeybindings()

	vp := viewport.New(0, 0)
	jvp := viewport.New(0, 0)
	h := help.New()

	return Model{
		store:        store,
		journalStore: journalStore,
		focusStore:   focusStore,
		cfg:          cfg,
		panelRatio:   cfg.PanelRatio,
		list:         l,
		journalList:  jl,
		viewport:     vp,
		journalViewport: jvp,
		help:         h,
		activeTab:    1, // Default to Active tab
	}
}

// Init returns the initial commands that load tasks and journal notes.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			tasks, err := m.store.List()
			return tasksLoaded{tasks: tasks, err: err}
		},
		func() tea.Msg {
			notes, err := m.journalStore.ListNotes(false)
			return notesLoaded{notes: notes, err: err}
		},
	)
}

// Update handles all incoming messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Forms need ALL message types (cursor blink, timers, etc.), not just KeyMsg.
	if m.mode == modeAdd || m.mode == modeEdit || m.mode == modeSubtask || m.mode == modeEditSubtask {
		return m.updateFormMsg(msg)
	}
	if m.mode == modeJournalAdd || m.mode == modeJournalEdit {
		return m.updateJournalForm(msg)
	}
	if m.mode == modeExport {
		return m.updateExportForm(msg)
	}
	if m.mode == modeTimeLog {
		return m.updateTimeLogForm(msg)
	}

	switch msg := msg.(type) {
	case tasksLoaded:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
			return m, nil
		}
		m.tasks = msg.tasks
		m.refreshList()
		m.updateDetail()
		return m, nil

	case notesLoaded:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
			return m, nil
		}
		m.notes = msg.notes
		m.refreshJournalList()
		m.updateJournalDetail()
		return m, nil

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case focusTickMsg:
		if !m.focusActive || m.focusSession == nil {
			return m, nil
		}
		now := time.Now()
		remaining := m.focusSession.Remaining(now)
		if remaining <= 0 {
			// Session complete.
			if err := m.focusStore.Complete(m.focusSession.ID); err != nil {
				return m, m.setError(err)
			}
			m.focusActive = false
			m.focusSession = nil
			return m, m.setStatus("Focus session complete!")
		}
		return m, focusTick()

	case exportDoneMsg:
		if msg.err != nil {
			return m, m.setError(msg.err)
		}
		return m, m.setStatus("Exported to " + msg.path)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeComponents()
		m.updateDetail()
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeConfirmDelete {
			return m.updateConfirmDelete(msg)
		}
		if m.mode == modeConfirmDeleteSubtask {
			return m.updateConfirmDeleteSubtask(msg)
		}
		if m.mode == modeJournalConfirmHide {
			return m.updateJournalConfirmHide(msg)
		}
		if m.mode == modeJournalConfirmDelete {
			return m.updateJournalConfirmDelete(msg)
		}
		if m.mode == modeFocusConfirmCancel {
			return m.updateFocusConfirmCancel(msg)
		}
		if m.mode == modeTagFilter {
			return m.updateTagFilter(msg)
		}
		if m.mode == modeBlockerPicker {
			return m.updateBlockerPicker(msg)
		}
		if m.mode == modeHelp {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
				m.mode = modeNormal
				return m, nil
			}
			return m, nil
		}
		if m.mode == modeStats {
			if msg.String() == "esc" || msg.String() == "G" || msg.String() == "q" {
				m.mode = modeNormal
				return m, nil
			}
			return m, nil
		}

		// Global keys handled before per-tab dispatch.
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.mode = modeHelp
			return m, nil
		case key.Matches(msg, keys.Tab):
			m.switchTab()
			return m, nil
		case key.Matches(msg, keys.Undo):
			return m.handleUndo()
		case key.Matches(msg, keys.Export):
			m.exportFormData = &ui.ExportFormData{Format: "md"}
			m.form = ui.ExportForm(m.exportFormData)
			m.mode = modeExport
			return m, m.form.Init()
		case key.Matches(msg, keys.Focus):
			return m.handleFocusToggle()
		case key.Matches(msg, keys.Stats):
			m.computeStats()
			m.mode = modeStats
			return m, nil
		case key.Matches(msg, keys.PanelWider):
			if m.panelRatio < 0.8 {
				m.panelRatio += 0.05
				m.resizeComponents()
				m.updateDetail()
			}
			return m, nil
		case key.Matches(msg, keys.PanelNarrower):
			if m.panelRatio > 0.2 {
				m.panelRatio -= 0.05
				m.resizeComponents()
				m.updateDetail()
			}
			return m, nil
		case key.Matches(msg, keys.TagBar):
			if m.activeTab != 3 {
				m.tagBarVisible = !m.tagBarVisible
				m.activeTag = ""
				m.resizeComponents()
				m.refreshList()
				m.updateDetail()
			}
			return m, nil
		}

		// Journal tab gets its own handler.
		if m.activeTab == 3 {
			return m.updateJournal(msg)
		}

		// While filtering, only handle Escape; delegate everything else to list.
		if m.list.FilterState() == list.Filtering {
			if key.Matches(msg, keys.Escape) {
				m.list.ResetFilter()
				m.list.SetShowFilter(false)
				return m, nil
			}
			break // fall through to list.Update() for filter input
		}

		// Dismiss applied filter with Escape, or return focus to list panel.
		if key.Matches(msg, keys.Escape) {
			if m.focusedPanel == 1 {
				m.focusedPanel = 0
				m.updateDetail()
				return m, nil
			}
			if m.list.FilterState() == list.FilterApplied {
				m.list.ResetFilter()
				m.list.SetShowFilter(false)
				return m, nil
			}
		}

		// Panel focus switching (lazygit-style).
		switch msg.String() {
		case "1":
			m.focusedPanel = 0
			m.updateDetail()
			return m, nil
		case "2":
			m.focusedPanel = 1
			m.updateDetail()
			return m, nil
		}

		// When detail panel is focused, keys operate on subtasks.
		if m.focusedPanel == 1 {
			selected := m.selectedTask()
			switch msg.String() {
			case "j", "down":
				if selected != nil && len(selected.Subtasks) > 0 {
					if m.subtaskIdx < len(selected.Subtasks)-1 {
						m.subtaskIdx++
					}
					m.updateDetail()
				}
				return m, nil
			case "k", "up":
				if m.subtaskIdx > 0 {
					m.subtaskIdx--
					m.updateDetail()
				}
				return m, nil
			}
			// Context-sensitive: a/e/d/s operate on subtasks when detail is focused.
			switch {
			case key.Matches(msg, keys.Add):
				if selected != nil {
					m.formData = &ui.TaskFormData{}
					m.form = ui.SubtaskForm(&m.formData.Title)
					m.mode = modeSubtask
					return m, m.form.Init()
				}
				return m, nil
			case key.Matches(msg, keys.Edit):
				if selected != nil && m.subtaskIdx >= 0 && m.subtaskIdx < len(selected.Subtasks) {
					st := selected.Subtasks[m.subtaskIdx]
					m.formData = &ui.TaskFormData{Title: st.Title}
					m.form = ui.SubtaskForm(&m.formData.Title)
					m.mode = modeEditSubtask
					return m, m.form.Init()
				}
				return m, nil
			case key.Matches(msg, keys.Delete):
				if selected != nil && m.subtaskIdx >= 0 && m.subtaskIdx < len(selected.Subtasks) {
					m.mode = modeConfirmDeleteSubtask
				}
				return m, nil
			case key.Matches(msg, keys.Status):
				if selected != nil && m.subtaskIdx >= 0 && m.subtaskIdx < len(selected.Subtasks) {
					st := selected.Subtasks[m.subtaskIdx]
					if err := m.store.ToggleSubtask(st.ID); err != nil {
						return m, m.setError(err)
					}
					if err := m.reload(); err != nil {
						return m, m.setError(err)
					}
					return m, m.setStatus("Subtask toggled")
				}
				return m, nil
			case key.Matches(msg, keys.TimeLog):
				if selected != nil {
					m.timeLogFormData = &ui.TimeLogFormData{}
					m.form = ui.TimeLogForm(m.timeLogFormData)
					m.mode = modeTimeLog
					return m, m.form.Init()
				}
				return m, nil
			case key.Matches(msg, keys.Blocker):
				if selected != nil {
					m.mode = modeBlockerPicker
				}
				return m, nil
			}
			// Other keys (q, ?, Tab, /, F1-F3) fall through to normal handling.
		}

		// Normal mode keybindings (list panel focused).
		switch {
		case key.Matches(msg, keys.Add):
			m.formData = &ui.TaskFormData{Priority: task.Medium, RecurFreq: "none"}
			m.form = ui.NewTaskForm(m.formData)
			m.mode = modeAdd
			return m, m.form.Init()

		case key.Matches(msg, keys.Edit):
			selected := m.selectedTask()
			if selected == nil {
				return m, nil
			}
			dueStr := ""
			if selected.DueDate != nil {
				dueStr = selected.DueDate.Format(time.DateOnly)
			}
			recurFreq := selected.RecurFreq.String()
			m.formData = &ui.TaskFormData{
				Title:       selected.Title,
				Description: selected.Description,
				Priority:    selected.Priority,
				DueDate:     dueStr,
				Tags:        strings.Join(selected.Tags, ", "),
				RecurFreq:   recurFreq,
			}
			m.form = ui.EditTaskForm(m.formData)
			m.mode = modeEdit
			return m, m.form.Init()

		case key.Matches(msg, keys.Delete):
			if m.selectedTask() != nil {
				m.mode = modeConfirmDelete
			}
			return m, nil

		case key.Matches(msg, keys.Status):
			selected := m.selectedTask()
			if selected != nil {
				oldStatus := selected.Status
				selected.Status = selected.Status.Next()

				// Handle recurring tasks: when completing, create next occurrence.
				if selected.Status == task.Done && selected.RecurFreq != task.RecurNone {
					nextDue := task.NextDueDate(*selected)
					newTask := &task.Task{
						Title:         selected.Title,
						Description:   selected.Description,
						Priority:      selected.Priority,
						DueDate:       &nextDue,
						Tags:          selected.Tags,
						RecurFreq:     selected.RecurFreq,
						RecurInterval: selected.RecurInterval,
					}
					if err := m.store.Create(newTask); err != nil {
						return m, m.setError(err)
					}
				}

				if err := m.store.Update(selected); err != nil {
					return m, m.setError(err)
				}
				// Capture undo for status change.
				taskID := selected.ID
				m.undoAction = &undoAction{
					description: fmt.Sprintf("Undo status change on %q", selected.Title),
					undo: func() error {
						t, err := m.store.GetByID(taskID)
						if err != nil {
							return err
						}
						t.Status = oldStatus
						return m.store.Update(t)
					},
				}
				if err := m.reload(); err != nil {
					return m, m.setError(err)
				}
				return m, m.setStatus("Status: " + selected.Status.String())
			}
			return m, nil

		case key.Matches(msg, keys.Subtask):
			if m.selectedTask() != nil {
				m.formData = &ui.TaskFormData{}
				m.form = ui.SubtaskForm(&m.formData.Title)
				m.mode = modeSubtask
				return m, m.form.Init()
			}
			return m, nil

		case key.Matches(msg, keys.SortDate):
			m.sortBy = sortCreated
			m.sortTasks()
			return m, nil

		case key.Matches(msg, keys.SortDue):
			m.sortBy = sortDue
			m.sortTasks()
			return m, nil

		case key.Matches(msg, keys.SortPrio):
			m.sortBy = sortPriority
			m.sortTasks()
			return m, nil

		case key.Matches(msg, keys.Search):
			m.list.SetShowFilter(true)
			m.list.SetFilteringEnabled(true)
			return m, nil
		}
	}

	// Delegate to list for navigation and filtering.
	var cmd tea.Cmd
	prevIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIndex {
		m.subtaskIdx = 0
		m.updateDetail()
	}

	return m, cmd
}

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

// View renders the entire application UI.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Calculate counts.
	allCount := len(m.tasks)
	var activeCount, doneCount int
	for _, t := range m.tasks {
		switch t.Status {
		case task.Done:
			doneCount++
		default:
			activeCount++
		}
	}

	// Header tabs.
	journalCount := len(m.notes)
	header := ui.RenderTabs(m.activeTab, allCount, activeCount, doneCount, journalCount, m.width)

	// Journal tab renders its own content.
	if m.activeTab == 3 {
		view := m.viewJournal(header)
		return m.renderJournalOverlays(view)
	}

	// Tag bar (optional, between header and content).
	var tagBar string
	if m.tagBarVisible {
		tagBar = ui.RenderTagBar(m.allTags(), m.activeTag, m.width)
	}

	// Content area height.
	contentHeight := m.height - lipgloss.Height(header) - 1 // 1 for status bar
	if m.tagBarVisible {
		contentHeight -= lipgloss.Height(tagBar)
	}

	// List panel.
	listWidth := int(float64(m.width) * m.panelRatio)
	detailWidth := m.width - listWidth

	var listContent string
	if len(m.list.Items()) == 0 {
		var emptyText string
		switch m.activeTab {
		case 1:
			if allCount == 0 {
				emptyText = "No tasks yet\n\nPress 'a' to add your first task"
			} else {
				emptyText = "No active tasks\n\nAll tasks are completed!"
			}
		case 2:
			emptyText = "No completed tasks yet"
		default:
			emptyText = "No tasks yet\n\nPress 'a' to add your first task\nPress '?' for help"
		}
		listContent = lipgloss.NewStyle().
			Foreground(gray).
			Align(lipgloss.Center).
			Width(listWidth - 4).
			Render("\n\n" + emptyText)
	} else {
		listContent = m.list.View()
	}

	// Panels with title in border (lazygit-style).
	listPanel := renderPanel(listContent, "1: Tasks", listWidth, contentHeight, m.focusedPanel == 0)
	detailPanel := renderPanel(m.viewport.View(), "2: Details", detailWidth, contentHeight, m.focusedPanel == 1)

	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)

	// Focus timer string.
	timerStr := m.focusTimerStr()

	// Status bar.
	statusBar := ui.RenderStatusBar(allCount, doneCount, activeCount, m.width, m.statusMsg, m.focusedPanel, m.activeTab, timerStr, m.undoAction != nil)

	// Combine all sections.
	var sections []string
	sections = append(sections, header)
	if m.tagBarVisible {
		sections = append(sections, tagBar)
	}
	sections = append(sections, content, statusBar)
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlay dialogs.
	switch m.mode {
	case modeAdd, modeEdit:
		if m.form != nil {
			title := "New Task"
			if m.mode == modeEdit {
				title = "Edit Task"
			}
			formView := m.form.View()
			dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render(title) + "\n\n" + formView
			dialog := dialogStyle.Render(dialogContent)
			view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeSubtask, modeEditSubtask:
		if m.form != nil {
			title := "Add Subtask"
			if m.mode == modeEditSubtask {
				title = "Edit Subtask"
			}
			formView := m.form.View()
			dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render(title) + "\n\n" + formView
			dialog := dialogStyle.Render(dialogContent)
			view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeExport:
		if m.form != nil {
			formView := m.form.View()
			dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render("Export") + "\n\n" + formView
			dialog := dialogStyle.Render(dialogContent)
			view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeTimeLog:
		if m.form != nil {
			formView := m.form.View()
			dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render("Log Time") + "\n\n" + formView
			dialog := dialogStyle.Render(dialogContent)
			view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeConfirmDelete:
		selected := m.selectedTask()
		title := ""
		if selected != nil {
			title = selected.Title
		}
		dialog := ui.RenderConfirmDialogBox("Delete Task?", fmt.Sprintf("Delete \"%s\"?", title))
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeConfirmDeleteSubtask:
		selected := m.selectedTask()
		stTitle := ""
		if selected != nil && m.subtaskIdx >= 0 && m.subtaskIdx < len(selected.Subtasks) {
			stTitle = selected.Subtasks[m.subtaskIdx].Title
		}
		dialog := ui.RenderConfirmDialogBox("Delete Subtask?", fmt.Sprintf("Delete \"%s\"?", stTitle))
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeFocusConfirmCancel:
		remaining := ""
		if m.focusSession != nil {
			remaining = focus.FormatTimer(m.focusSession.Remaining(time.Now()))
		}
		dialog := ui.RenderConfirmDialogBox("Cancel Focus?", fmt.Sprintf("Cancel session with %s remaining?", remaining), ui.Yellow)
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeTagFilter:
		// Tag filter uses the tag bar + inline selection, no overlay needed.

	case modeBlockerPicker:
		dialog := m.renderBlockerOverlay()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeStats:
		statsView := m.renderStatsOverlay()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, statsView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeHelp:
		helpView := m.renderHelpOverlay()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
	}

	return view
}

func (m *Model) resizeComponents() {
	header := ui.RenderTabs(m.activeTab, 0, 0, 0, 0, m.width)
	headerHeight := lipgloss.Height(header)
	statusBarHeight := 1
	tagBarHeight := 0
	if m.tagBarVisible {
		tagBarHeight = 1
	}
	contentHeight := m.height - headerHeight - statusBarHeight - tagBarHeight

	listWidth := int(float64(m.width) * m.panelRatio)
	detailWidth := m.width - listWidth

	// Both panels: border(2w,2h) + padding(0,1)=(2w,0h) → frame: 4w, 2h
	m.list.SetSize(listWidth-4, contentHeight-2)
	m.viewport.Width = detailWidth - 4
	m.viewport.Height = contentHeight - 2

	// Journal components share the same layout dimensions.
	m.journalList.SetSize(listWidth-4, contentHeight-2)
	m.journalViewport.Width = detailWidth - 4
	m.journalViewport.Height = contentHeight - 2

	m.help.Width = m.width
}

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

func (m *Model) switchTab() {
	m.activeTab = (m.activeTab + 1) % tabCount
	m.focusedPanel = 0
	m.subtaskIdx = 0
	m.entryIdx = 0
	if m.activeTab != 3 {
		m.refreshList()
		m.updateDetail()
	} else {
		m.refreshJournalList()
		m.updateJournalDetail()
	}
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

func (m *Model) focusTimerStr() string {
	if !m.focusActive || m.focusSession == nil {
		return ""
	}
	remaining := m.focusSession.Remaining(time.Now())
	return "🍅 " + focus.FormatTimer(remaining)
}

func (m *Model) computeStats() {
	s := &statsData{
		totalTasks: len(m.tasks),
		tagCounts:  make(map[string]int),
	}
	for _, t := range m.tasks {
		switch t.Status {
		case task.Done:
			s.doneTasks++
		default:
			s.activeTasks++
		}
		switch t.Priority {
		case task.Low:
			s.lowCount++
		case task.Medium:
			s.medCount++
		case task.High:
			s.highCount++
		case task.Urgent:
			s.urgentCount++
		}
		for _, tag := range t.Tags {
			s.tagCounts[tag]++
		}
	}
	if m.focusStore != nil {
		s.focusToday, _ = m.focusStore.TodayCount()
	}
	if m.journalStore != nil {
		completions := make(map[string]int)
		for _, n := range m.notes {
			if len(n.Entries) > 0 {
				completions[n.Date.Format(time.DateOnly)] = len(n.Entries)
			}
		}
		s.journalStreak = ui.RenderJournalStreak(completions, 30)
	}
	m.stats = s
}

func (m Model) renderStatsOverlay() string {
	if m.stats == nil {
		return ""
	}
	s := m.stats

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Statistics"))
	lines = append(lines, "")

	// Task counts.
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Tasks"))
	lines = append(lines, fmt.Sprintf("  Total: %d  Active: %d  Done: %d", s.totalTasks, s.activeTasks, s.doneTasks))
	lines = append(lines, "")

	// Priority breakdown.
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Priority"))
	lines = append(lines, "  "+ui.RenderPriorityBreakdown(s.lowCount, s.medCount, s.highCount, s.urgentCount))
	lines = append(lines, "")

	// Tags.
	if len(s.tagCounts) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Tags"))
		lines = append(lines, "  "+ui.RenderTagCloud(s.tagCounts))
		lines = append(lines, "")
	}

	// Focus sessions today.
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Focus"))
	lines = append(lines, fmt.Sprintf("  Sessions today: %d", s.focusToday))
	lines = append(lines, "")

	// Journal streak.
	if s.journalStreak != "" {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Journal"))
		lines = append(lines, "  "+s.journalStreak)
		lines = append(lines, "")
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc or G to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 3).
		Width(60).
		Render(content)
}

func (m Model) renderBlockerOverlay() string {
	selected := m.selectedTask()
	if selected == nil {
		return ""
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Task Dependencies"))
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("Task: %s", selected.Title)))
	lines = append(lines, "")

	if len(selected.BlockedByIDs) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("No blockers"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render("Blocked by:"))
		for _, id := range selected.BlockedByIDs {
			blocker, err := m.store.GetByID(id)
			if err != nil {
				lines = append(lines, fmt.Sprintf("  - Task #%d (not found)", id))
				continue
			}
			statusIcon := blocker.Status.Icon()
			style := lipgloss.NewStyle().Foreground(ui.White)
			if blocker.Status == task.Done {
				style = style.Foreground(ui.Green)
			}
			lines = append(lines, style.Render(fmt.Sprintf("  %s %s", statusIcon, blocker.Title)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 2).
		Width(50).
		Render(content)
}

// renderPanel draws a bordered panel with the title embedded in the top border (lazygit-style).
func renderPanel(content, title string, width, height int, focused bool) string {
	if width < 4 || height < 3 {
		return content
	}

	borderColor := gray
	if focused {
		borderColor = cyan
	}

	bc := lipgloss.NewStyle().Foreground(borderColor)
	tc := lipgloss.NewStyle().Foreground(borderColor).Bold(focused)

	border := lipgloss.RoundedBorder()
	innerWidth := width - 2

	// Pad and size the content area.
	padded := lipgloss.NewStyle().
		Width(innerWidth).
		Height(height - 2).
		PaddingLeft(1).
		PaddingRight(1).
		Render(content)

	// Top border with title: ╭─ 1 Tasks ──────────────╮
	titleRendered := tc.Render(title)
	titleVisualWidth := lipgloss.Width(titleRendered)
	fillWidth := innerWidth - titleVisualWidth - 3 // "─ " before + " " after
	if fillWidth < 0 {
		fillWidth = 0
	}
	topLine := bc.Render(border.TopLeft+border.Top+" ") +
		titleRendered +
		bc.Render(" "+strings.Repeat(border.Top, fillWidth)+border.TopRight)

	// Bottom border: ╰──────────────╯
	bottomLine := bc.Render(border.BottomLeft + strings.Repeat(border.Bottom, innerWidth) + border.BottomRight)

	// Add side borders to each content line.
	lines := strings.Split(padded, "\n")
	result := make([]string, 0, len(lines)+2)
	result = append(result, topLine)
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > innerWidth {
			line = lipgloss.NewStyle().MaxWidth(innerWidth).Render(line)
			lineWidth = lipgloss.Width(line)
		}
		if lineWidth < innerWidth {
			line += strings.Repeat(" ", innerWidth-lineWidth)
		}
		result = append(result, bc.Render(border.Left)+line+bc.Render(border.Right))
	}
	result = append(result, bottomLine)

	return strings.Join(result, "\n")
}

func (m Model) renderHelpOverlay() string {
	helpLines := []struct{ key, desc string }{
		{"", "Panels:"},
		{"1", "Focus task list"},
		{"2", "Focus detail panel"},
		{"Esc", "Back to list / clear filter"},
		{"j/k", "Navigate items in panel"},
		{"Tab", "Switch tab"},
		{"< / >", "Resize panels"},
		{"", ""},
		{"", "1: Tasks (list focused):"},
		{"a", "Add new task"},
		{"e", "Edit task"},
		{"d", "Delete task"},
		{"s", "Cycle task status"},
		{"/", "Search / filter"},
		{"F1/F2/F3", "Sort date/due/prio"},
		{"F4", "Toggle tag filter bar"},
		{"", ""},
		{"", "2: Details (detail focused):"},
		{"a", "Add subtask"},
		{"e", "Edit subtask"},
		{"d", "Delete subtask"},
		{"s", "Toggle subtask"},
		{"l", "Log time"},
		{"b", "View blockers"},
		{"", ""},
		{"", "Global:"},
		{"p", "Start/cancel focus timer"},
		{"X", "Export tasks"},
		{"G", "Statistics dashboard"},
		{"Ctrl+Z", "Undo last action"},
		{"", ""},
		{"", "Forms:"},
		{"Tab/S-Tab", "Next / prev field"},
		{"Enter", "Submit"},
		{"Esc", "Cancel"},
		{"", ""},
		{"?", "Toggle this help"},
		{"q", "Quit"},
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Keyboard Shortcuts"))
	lines = append(lines, "")
	for _, h := range helpLines {
		if h.key == "" && h.desc == "" {
			lines = append(lines, "")
			continue
		}
		if h.key == "" {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(cyan).Render(h.desc))
			continue
		}
		k := lipgloss.NewStyle().Foreground(cyan).Width(16).Render(h.key)
		d := lipgloss.NewStyle().Foreground(gray).Render(h.desc)
		lines = append(lines, k+d)
	}
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(gray).Render("Press Esc or ? to close"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyan).
		Padding(1, 3).
		Width(50).
		Render(content)
}
