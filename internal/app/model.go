package app

import (
	"fmt"
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

		}
	}

	// Delegate to the active list for navigation and filtering.
	// This also handles internal list messages (e.g. filter matching) that
	// are not tea.KeyMsg and fall through the type switch above.
	var cmd tea.Cmd
	if m.activeTab == 3 {
		prevIndex := m.journalList.Index()
		m.journalList, cmd = m.journalList.Update(msg)
		if m.journalList.Index() != prevIndex {
			m.entryIdx = 0
			m.updateJournalDetail()
		}
		m.journalList.SetShowFilter(m.journalList.FilterState() == list.Filtering || m.journalList.FilterState() == list.FilterApplied)
	} else {
		prevIndex := m.list.Index()
		m.list, cmd = m.list.Update(msg)
		if m.list.Index() != prevIndex {
			m.subtaskIdx = 0
			m.updateDetail()
		}
		m.list.SetShowFilter(m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied)
	}

	return m, cmd
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
	if len(m.list.Items()) == 0 && m.list.FilterState() == list.Unfiltered {
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







