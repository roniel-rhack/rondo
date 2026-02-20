package app

import (
	"fmt"
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
)

const tabCount = 4

type sortOrder int

const (
	sortCreated sortOrder = iota
	sortDue
	sortPriority
)

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
func New(store *task.Store, journalStore *journal.Store) Model {
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
		store:           store,
		journalStore:    journalStore,
		list:            l,
		journalList:     jl,
		viewport:        vp,
		journalViewport: jvp,
		help:            h,
		activeTab:       1, // Default to Active tab
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
		if m.mode == modeHelp {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
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
			}
			// Other keys (q, ?, Tab, /, F1-F3) fall through to normal handling.
		}

		// Normal mode keybindings (list panel focused).
		switch {
		case key.Matches(msg, keys.Add):
			m.formData = &ui.TaskFormData{Priority: task.Medium}
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
			m.formData = &ui.TaskFormData{
				Title:       selected.Title,
				Description: selected.Description,
				Priority:    selected.Priority,
				DueDate:     dueStr,
				Tags:        strings.Join(selected.Tags, ", "),
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
				selected.Status = selected.Status.Next()
				if err := m.store.Update(selected); err != nil {
					return m, m.setError(err)
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

func (m *Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		selected := m.selectedTask()
		if selected != nil {
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

	var statusCmd tea.Cmd
	if m.mode == modeAdd {
		t := &task.Task{
			Title:       m.formData.Title,
			Description: m.formData.Description,
			Priority:    m.formData.Priority,
			DueDate:     dueDate,
			Tags:        tags,
		}
		if err := m.store.Create(t); err != nil {
			statusCmd = m.setError(err)
		} else {
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
			if err := m.store.Update(selected); err != nil {
				statusCmd = m.setError(err)
			} else {
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
	// Errors are shown inline via statusMsg, not here.

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

	// Content area height.
	contentHeight := m.height - lipgloss.Height(header) - 1 // 1 for status bar

	// List panel (40% width).
	listWidth := m.width * 2 / 5
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

	// Status bar.
	statusBar := ui.RenderStatusBar(allCount, doneCount, activeCount, m.width, m.statusMsg, m.focusedPanel, m.activeTab)

	// Combine all sections.
	view := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)

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

	case modeHelp:
		helpView := m.renderHelpOverlay()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
	}

	// Note: journal tab overlays are handled by renderJournalOverlays in viewJournal path.

	return view
}

func (m *Model) resizeComponents() {
	header := ui.RenderTabs(m.activeTab, 0, 0, 0, 0, m.width)
	headerHeight := lipgloss.Height(header)
	statusBarHeight := 1
	contentHeight := m.height - headerHeight - statusBarHeight

	listWidth := m.width * 2 / 5
	detailWidth := m.width - listWidth

	// Both panels: border(2w,2h) + padding(0,1)=(2w,0h) → frame: 4w, 2h
	// Title is embedded in the border, no extra line needed.
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
	if m.activeTab == 0 {
		return m.tasks
	}
	var filtered []task.Task
	for _, t := range m.tasks {
		switch m.activeTab {
		case 1: // Active (Pending + InProgress)
			if t.Status != task.Done {
				filtered = append(filtered, t)
			}
		case 2: // Done
			if t.Status == task.Done {
				filtered = append(filtered, t)
			}
		}
	}
	return filtered
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
		{"", ""},
		{"", "1: Tasks (list focused):"},
		{"a", "Add new task"},
		{"e", "Edit task"},
		{"d", "Delete task"},
		{"s", "Cycle task status"},
		{"/", "Search / filter"},
		{"F1/F2/F3", "Sort date/due/prio"},
		{"", ""},
		{"", "2: Details (detail focused):"},
		{"a", "Add subtask"},
		{"e", "Edit subtask"},
		{"d", "Delete subtask"},
		{"s", "Toggle subtask"},
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
