package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/ui"
)

// notesLoaded is a message sent after journal notes are loaded.
type notesLoaded struct {
	notes []journal.Note
	err   error
}

// updateJournal handles KeyMsg when activeTab == 3.
func (m *Model) updateJournal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// While filtering journal list, only handle Escape.
	if m.journalList.FilterState() == list.Filtering {
		if key.Matches(msg, keys.Escape) {
			m.journalList.ResetFilter()
			m.journalList.SetShowFilter(false)
			return m, nil
		}
		// Fall through to journalList.Update() below.
		var cmd tea.Cmd
		m.journalList, cmd = m.journalList.Update(msg)
		m.journalList.SetShowFilter(m.journalList.FilterState() == list.Filtering || m.journalList.FilterState() == list.FilterApplied)
		return m, cmd
	}

	// Escape: return focus to list or clear filter.
	if key.Matches(msg, keys.Escape) {
		if m.focusedPanel == 1 {
			m.focusedPanel = 0
			m.entryIdx = 0
			m.updateJournalDetail()
			return m, nil
		}
		if m.journalList.FilterState() == list.FilterApplied {
			m.journalList.ResetFilter()
			m.journalList.SetShowFilter(false)
			return m, nil
		}
	}

	// Panel focus switching.
	switch msg.String() {
	case "1":
		m.focusedPanel = 0
		m.updateJournalDetail()
		return m, nil
	case "2":
		m.focusedPanel = 1
		m.updateJournalDetail()
		return m, nil
	}

	// Add entry: always adds to today's note.
	if key.Matches(msg, keys.Add) {
		m.journalFormData = &ui.JournalFormData{}
		m.form = ui.JournalEntryForm(&m.journalFormData.Body)
		m.mode = modeJournalAdd
		return m, m.form.Init()
	}

	// Keys only active on panel 0 (notes list).
	if m.focusedPanel == 0 {
		switch {
		case key.Matches(msg, keys.Hide):
			if m.selectedNote() != nil {
				m.mode = modeJournalConfirmHide
			}
			return m, nil
		case key.Matches(msg, keys.ShowHidden):
			m.showHidden = !m.showHidden
			if err := m.reloadJournal(); err != nil {
				return m, m.setError(err)
			}
			if m.showHidden {
				return m, m.setStatus("Showing hidden notes")
			}
			return m, m.setStatus("Hiding hidden notes")
		}
	}

	// Entries panel focused: cursor navigation + edit/delete.
	if m.focusedPanel == 1 {
		note := m.selectedNote()
		entryCount := 0
		if note != nil {
			entryCount = len(note.Entries)
		}

		switch msg.String() {
		case "j", "down":
			if m.entryIdx < entryCount-1 {
				m.entryIdx++
				m.updateJournalDetail()
			}
			return m, nil
		case "k", "up":
			if m.entryIdx > 0 {
				m.entryIdx--
				m.updateJournalDetail()
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, keys.Edit):
			if note != nil && m.entryIdx >= 0 && m.entryIdx < entryCount {
				entry := note.Entries[m.entryIdx]
				m.journalFormData = &ui.JournalFormData{Body: entry.Body}
				m.form = ui.JournalEntryForm(&m.journalFormData.Body)
				m.mode = modeJournalEdit
				return m, m.form.Init()
			}
			return m, nil
		case key.Matches(msg, keys.Delete):
			if note != nil && m.entryIdx >= 0 && m.entryIdx < entryCount {
				m.mode = modeJournalConfirmDelete
			}
			return m, nil
		}

		// Fall through to viewport for other keys (PgUp, PgDn, etc.).
		var cmd tea.Cmd
		m.journalViewport, cmd = m.journalViewport.Update(msg)
		return m, cmd
	}

	// Notes list panel: delegate for navigation.
	var cmd tea.Cmd
	prevIndex := m.journalList.Index()
	m.journalList, cmd = m.journalList.Update(msg)
	if m.journalList.Index() != prevIndex {
		m.entryIdx = 0
		m.updateJournalDetail()
	}

	// Show/hide filter bar based on filtering state.
	m.journalList.SetShowFilter(m.journalList.FilterState() == list.Filtering || m.journalList.FilterState() == list.FilterApplied)

	return m, cmd
}

// updateJournalForm handles all messages for the journal entry form.
func (m *Model) updateJournalForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.resizeComponents()
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.mode = modeNormal
		m.form = nil
		m.journalFormData = nil
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			body := ""
			if m.journalFormData != nil {
				body = strings.TrimSpace(m.journalFormData.Body)
			}
			currentMode := m.mode
			m.mode = modeNormal
			m.form = nil
			m.journalFormData = nil
			if body == "" {
				return m, nil
			}

			if currentMode == modeJournalEdit {
				note := m.selectedNote()
				if note != nil && m.entryIdx >= 0 && m.entryIdx < len(note.Entries) {
					entry := note.Entries[m.entryIdx]
					if err := m.journalStore.UpdateEntry(entry.ID, body); err != nil {
						return m, m.setError(err)
					}
				}
				if err := m.reloadJournal(); err != nil {
					return m, m.setError(err)
				}
				return m, m.setStatus("Entry updated")
			}

			// modeJournalAdd
			note, err := m.journalStore.GetOrCreateToday()
			if err != nil {
				return m, m.setError(err)
			}
			if err := m.journalStore.AddEntry(note.ID, body); err != nil {
				return m, m.setError(err)
			}
			if err := m.reloadJournal(); err != nil {
				return m, m.setError(err)
			}
			return m, m.setStatus("Entry added")
		}
		if m.form.State == huh.StateAborted {
			m.mode = modeNormal
			m.form = nil
			m.journalFormData = nil
			return m, nil
		}
	}
	return m, cmd
}

// updateJournalConfirmHide handles the hide/unhide confirmation dialog.
func (m *Model) updateJournalConfirmHide(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		note := m.selectedNote()
		if note == nil {
			m.mode = modeNormal
			return m, m.setStatus("No note selected")
		}
		wasHidden := note.Hidden
		if err := m.journalStore.ToggleHidden(note.ID); err != nil {
			m.mode = modeNormal
			return m, m.setError(err)
		}
		if err := m.reloadJournal(); err != nil {
			m.mode = modeNormal
			return m, m.setError(err)
		}
		m.mode = modeNormal
		if !wasHidden {
			return m, m.setStatus("Note hidden")
		}
		return m, m.setStatus("Note restored")
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

// updateJournalConfirmDelete handles the entry delete confirmation dialog.
func (m *Model) updateJournalConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		note := m.selectedNote()
		if note == nil || m.entryIdx < 0 || m.entryIdx >= len(note.Entries) {
			m.mode = modeNormal
			return m, m.setStatus("No entry selected")
		}
		entry := note.Entries[m.entryIdx]
		// Capture for undo.
		noteID := note.ID
		entryBody := entry.Body
		entryCreatedAt := entry.CreatedAt
		m.undoAction = &undoAction{
			description: "Undo delete journal entry",
			undo: func() error {
				return m.journalStore.RestoreEntry(noteID, entryBody, entryCreatedAt)
			},
		}
		if err := m.journalStore.DeleteEntry(entry.ID); err != nil {
			m.mode = modeNormal
			return m, m.setError(err)
		}
		if m.entryIdx > 0 && m.entryIdx >= len(note.Entries)-1 {
			m.entryIdx--
		}
		if err := m.reloadJournal(); err != nil {
			m.mode = modeNormal
			return m, m.setError(err)
		}
		m.mode = modeNormal
		return m, m.setStatus("Entry deleted")
	case "n", "N", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m *Model) selectedNote() *journal.Note {
	item := m.journalList.SelectedItem()
	if item == nil {
		return nil
	}
	n, ok := item.(journal.Note)
	if !ok {
		return nil
	}
	for i := range m.notes {
		if m.notes[i].ID == n.ID {
			return &m.notes[i]
		}
	}
	return nil
}

func (m *Model) reloadJournal() error {
	var selectedID int64
	if sel := m.selectedNote(); sel != nil {
		selectedID = sel.ID
	}

	notes, err := m.journalStore.ListNotes(m.showHidden)
	if err != nil {
		return err
	}
	m.notes = notes
	m.refreshJournalList()

	if selectedID != 0 {
		for i, item := range m.journalList.Items() {
			if n, ok := item.(journal.Note); ok && n.ID == selectedID {
				m.journalList.Select(i)
				break
			}
		}
	}
	// Clamp entryIdx after reload.
	note := m.selectedNote()
	if note == nil || len(note.Entries) == 0 {
		m.entryIdx = 0
	} else if m.entryIdx >= len(note.Entries) {
		m.entryIdx = len(note.Entries) - 1
	}
	m.updateJournalDetail()
	return nil
}

func (m *Model) refreshJournalList() {
	var filtered []journal.Note
	if m.showHidden {
		filtered = m.notes
	} else {
		for _, n := range m.notes {
			if !n.Hidden {
				filtered = append(filtered, n)
			}
		}
	}
	items := make([]list.Item, len(filtered))
	for i, n := range filtered {
		items[i] = n
	}
	m.journalList.SetItems(items)
}

func (m *Model) updateJournalDetail() {
	note := m.selectedNote()
	content := ui.RenderJournalDetail(note, m.journalViewport.Width, m.entryIdx, m.focusedPanel == 1)
	m.journalViewport.SetContent(content)
	m.journalViewport.GotoTop()
}

// viewJournal renders the journal tab content (both panels + status bar).
func (m Model) viewJournal(header string) string {
	contentHeight := m.height - lipgloss.Height(header) - 1

	listWidth := int(float64(m.width) * m.panelRatio)
	detailWidth := m.width - listWidth

	var listContent string
	if len(m.journalList.Items()) == 0 && m.journalList.FilterState() == list.Unfiltered {
		var emptyText string
		if len(m.notes) == 0 {
			emptyText = "No journal entries yet\n\nPress 'a' to write your\nfirst entry"
		} else {
			emptyText = "All notes are hidden\n\nPress 'H' to reveal them\nor 'a' to start a new one"
		}
		listContent = lipgloss.NewStyle().
			Foreground(gray).
			Align(lipgloss.Center).
			Width(listWidth - 4).
			Render("\n\n" + emptyText)
	} else {
		listContent = m.journalList.View()
	}

	listTitle := "1: Notes"
	if m.showHidden {
		listTitle = "1: Notes (all)"
	}

	// Dynamic detail panel title.
	detailTitle := "2: Journal"
	if note := m.selectedNote(); note != nil {
		detailTitle = "2: " + note.DateTitle()
	}

	listPanel := renderPanel(listContent, listTitle, listWidth, contentHeight, m.focusedPanel == 0)
	detailPanel := renderPanel(m.journalViewport.View(), detailTitle, detailWidth, contentHeight, m.focusedPanel == 1)

	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)

	// Journal status bar counts.
	noteCount := len(m.notes)
	todayEntryCount := 0
	today := todayDate()
	for _, n := range m.notes {
		if n.Date.Equal(today) {
			todayEntryCount = len(n.Entries)
			break
		}
	}
	timerStr := m.focusTimerStr()
	statusBar := ui.RenderStatusBar(noteCount, todayEntryCount, 0, m.width, m.statusMsg, m.focusedPanel, 3, timerStr, m.undoAction != nil)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

// renderJournalHelpOverlay renders the help overlay for journal context.
func (m Model) renderJournalHelpOverlay() string {
	helpLines := []struct{ key, desc string }{
		{"", "Navigation"},
		{"1 / 2", "Focus notes / entries panel"},
		{"j/k", "Navigate items"},
		{"Tab", "Switch tab"},
		{"< / >", "Resize panels"},
		{"Esc", "Back to list / clear filter"},
		{"", ""},
		{"", "Journal"},
		{"a", "Add entry (today)"},
		{"e / d", "Edit / delete entry (entries)"},
		{"h", "Hide / restore note"},
		{"H", "Toggle show hidden"},
		{"/", "Search notes"},
		{"", ""},
		{"", "Tools"},
		{"p", "Focus timer"},
		{"X", "Export"},
		{"G", "Statistics"},
		{"Ctrl+Z", "Undo"},
		{"?", "This help"},
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

// renderJournalOverlays renders journal-specific dialog overlays.
func (m Model) renderJournalOverlays(view string) string {
	switch m.mode {
	case modeJournalAdd, modeJournalEdit:
		if m.form != nil {
			title := "New Journal Entry"
			if m.mode == modeJournalEdit {
				title = "Edit Journal Entry"
			}
			formView := m.form.View()
			dialogContent := lipgloss.NewStyle().Bold(true).Foreground(white).Render(title) + "\n\n" + formView
			dialog := dialogStyle.Render(dialogContent)
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeJournalConfirmDelete:
		note := m.selectedNote()
		if note != nil && m.entryIdx >= 0 && m.entryIdx < len(note.Entries) {
			entry := note.Entries[m.entryIdx]
			preview := entry.Body
			if len(preview) > 40 {
				preview = preview[:40] + "..."
			}
			msg := fmt.Sprintf("Delete entry from %s?\n\"%s\"", entry.CreatedAt.Format("3:04 PM"), preview)
			dialog := ui.RenderConfirmDialogBox("Delete Entry?", msg)
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeJournalConfirmHide:
		note := m.selectedNote()
		if note != nil {
			action := "Hide"
			message := fmt.Sprintf("Hide \"%s\" (%d entries)?\nThe note can be restored later.", note.DateTitle(), len(note.Entries))
			borderColor := ui.Yellow
			if note.Hidden {
				action = "Restore"
				message = fmt.Sprintf("Restore \"%s\" (%d entries)?", note.DateTitle(), len(note.Entries))
				borderColor = ui.Cyan
			}
			dialog := ui.RenderConfirmDialogBox(action+" Note?", message, borderColor)
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		}

	case modeHelp:
		helpView := m.renderJournalHelpOverlay()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeStats:
		statsView := m.renderStatsOverlay()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, statsView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))

	case modeFocusConfirmCancel:
		remaining := ""
		if m.focusSession != nil {
			remaining = fmt.Sprintf("%s remaining", m.focusSession.Remaining(time.Now()).Round(time.Second))
		}
		dialog := ui.RenderConfirmDialogBox("Cancel Focus?", fmt.Sprintf("Cancel session with %s?", remaining), ui.Yellow)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
	}
	return view
}

func todayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}
