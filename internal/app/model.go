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

	"github.com/roniel/todo-app/internal/task"
	"github.com/roniel/todo-app/internal/ui"
)

type mode int

const (
	modeNormal mode = iota
	modeAdd
	modeEdit
	modeConfirmDelete
	modeSubtask
	modeHelp
)

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

	subtaskTitle string

	mode      mode
	activeTab int // 0=All, 1=Active, 2=Done
	sortBy    sortOrder
	width     int
	height    int
	ready     bool
	err       error
}

// tasksLoaded is a message sent after the initial data load completes.
type tasksLoaded struct {
	tasks []task.Task
	err   error
}

// New creates a new Model backed by the given store.
func New(store *task.Store) Model {
	delegate := newTaskDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(false)
	l.DisableQuitKeybindings()

	vp := viewport.New(0, 0)
	h := help.New()

	return Model{
		store:    store,
		list:     l,
		viewport: vp,
		help:     h,
	}
}

// Init returns the initial command that loads tasks from the store.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.store.List()
		return tasksLoaded{tasks: tasks, err: err}
	}
}

// Update handles all incoming messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoaded:
		m.tasks = msg.tasks
		if msg.err != nil {
			m.err = msg.err
		}
		m.refreshList()
		m.updateDetail()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeComponents()
		return m, nil

	case tea.KeyMsg:
		// Delegate to form when in form modes.
		if m.mode == modeAdd || m.mode == modeEdit {
			return m.updateForm(msg)
		}
		if m.mode == modeSubtask {
			return m.updateSubtaskForm(msg)
		}
		if m.mode == modeConfirmDelete {
			return m.updateConfirmDelete(msg)
		}
		if m.mode == modeHelp {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
				m.mode = modeNormal
				return m, nil
			}
			return m, nil
		}

		// Normal mode keybindings.
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.mode = modeHelp
			return m, nil

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
				m.store.Update(selected)
				m.reload()
			}
			return m, nil

		case key.Matches(msg, keys.Subtask):
			if m.selectedTask() != nil {
				m.subtaskTitle = ""
				m.form = ui.SubtaskForm(&m.subtaskTitle)
				m.mode = modeSubtask
				return m, m.form.Init()
			}
			return m, nil

		case key.Matches(msg, keys.ToggleSub):
			selected := m.selectedTask()
			if selected != nil {
				for _, st := range selected.Subtasks {
					if !st.Completed {
						m.store.ToggleSubtask(st.ID)
						m.reload()
						break
					}
				}
			}
			return m, nil

		case key.Matches(msg, keys.Tab):
			m.activeTab = (m.activeTab + 1) % 3
			m.refreshList()
			m.updateDetail()
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
		}
	}

	// Delegate to list for navigation and filtering.
	var cmd tea.Cmd
	prevIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIndex {
		m.updateDetail()
	}

	return m, cmd
}

func (m *Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		m.formData = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			m.submitTaskForm()
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) updateSubtaskForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			selected := m.selectedTask()
			if selected != nil && m.subtaskTitle != "" {
				m.store.AddSubtask(selected.ID, m.subtaskTitle)
			}
			m.mode = modeNormal
			m.form = nil
			m.reload()
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
			m.store.Delete(selected.ID)
			m.reload()
		}
		m.mode = modeNormal
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) submitTaskForm() {
	if m.formData == nil {
		return
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

	if m.mode == modeAdd {
		t := &task.Task{
			Title:       m.formData.Title,
			Description: m.formData.Description,
			Priority:    m.formData.Priority,
			DueDate:     dueDate,
			Tags:        tags,
		}
		m.store.Create(t)
	} else if m.mode == modeEdit {
		selected := m.selectedTask()
		if selected != nil {
			selected.Title = m.formData.Title
			selected.Description = m.formData.Description
			selected.Priority = m.formData.Priority
			selected.DueDate = dueDate
			selected.Tags = tags
			m.store.Update(selected)
		}
	}

	m.mode = modeNormal
	m.form = nil
	m.formData = nil
	m.reload()
}

// View renders the entire application UI.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
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
	inProgress := 0
	for _, t := range m.tasks {
		if t.Status == task.InProgress {
			inProgress++
		}
	}

	// Header tabs.
	header := ui.RenderTabs(m.activeTab, allCount, activeCount, doneCount, m.width)

	// Content area height.
	contentHeight := m.height - lipgloss.Height(header) - 1 // 1 for status bar

	// List panel (40% width).
	listWidth := m.width * 2 / 5
	detailWidth := m.width - listWidth

	listPanel := listPanelFocusedStyle.
		Width(listWidth - 2). // account for border
		Height(contentHeight - 2).
		Render(m.list.View())

	// Detail panel.
	detailContent := m.viewport.View()
	detailPanel := detailPanelStyle.
		Width(detailWidth - 2).
		Height(contentHeight - 2).
		Render(detailContent)

	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)

	// Status bar.
	statusBar := ui.RenderStatusBar(allCount, doneCount, inProgress, m.width)

	// Combine all sections.
	view := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)

	// Overlay dialogs.
	switch m.mode {
	case modeAdd, modeEdit:
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

	case modeSubtask:
		formView := m.form.View()
		dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render("Add Subtask") + "\n\n" + formView
		dialog := dialogStyle.Render(dialogContent)
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeConfirmDelete:
		selected := m.selectedTask()
		title := ""
		if selected != nil {
			title = selected.Title
		}
		view = ui.RenderConfirmDialog("Delete Task?", fmt.Sprintf("Delete \"%s\"?", title), m.width, m.height)

	case modeHelp:
		helpView := m.renderHelpOverlay()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
	}

	return view
}

func (m *Model) resizeComponents() {
	headerHeight := 3
	statusBarHeight := 1
	contentHeight := m.height - headerHeight - statusBarHeight

	listWidth := m.width*2/5 - 4 // borders + padding
	detailWidth := m.width - m.width*2/5 - 6

	m.list.SetSize(listWidth, contentHeight-2)
	m.viewport.Width = detailWidth
	m.viewport.Height = contentHeight - 4
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

func (m *Model) reload() {
	tasks, err := m.store.List()
	if err != nil {
		m.err = err
		return
	}
	m.tasks = tasks
	m.sortTasks()
	m.refreshList()
	m.updateDetail()
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
	detailWidth := m.width - m.width*2/5 - 6
	content := ui.RenderDetail(selected, detailWidth)
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

func (m Model) renderHelpOverlay() string {
	helpLines := []struct{ key, desc string }{
		{"j/k up/dn", "Navigate tasks"},
		{"Enter", "Select task"},
		{"a", "Add new task"},
		{"e", "Edit selected task"},
		{"d", "Delete selected task"},
		{"s", "Cycle task status"},
		{"t", "Add subtask"},
		{"x", "Toggle next subtask"},
		{"/", "Search / filter"},
		{"Tab", "Switch tab"},
		{"F1", "Sort by created date"},
		{"F2", "Sort by due date"},
		{"F3", "Sort by priority"},
		{"?", "Toggle this help"},
		{"q", "Quit"},
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(white).Render("Keyboard Shortcuts"))
	lines = append(lines, "")
	for _, h := range helpLines {
		k := lipgloss.NewStyle().Foreground(cyan).Width(14).Render(h.key)
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
		Width(45).
		Render(content)
}
