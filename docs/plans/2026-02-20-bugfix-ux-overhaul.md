# Bugfix & UX Overhaul Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all critical bugs (subtasks, tabs, layout) and improve TUI intuitiveness to professional standards.

**Architecture:** Fix the value/pointer receiver mismatch at the root, then fix layout calculations, form lifecycle, state management, and layer UX improvements on top.

**Tech Stack:** Go, Bubbletea v1.3, Bubbles v1.0, Lip Gloss v1.1, Huh v0.8

---

### Task 1: Fix value/pointer receiver — root cause of subtask bug

The `Update` method uses a value receiver, causing the model to be copied on every call. This breaks `SubtaskForm(&m.subtaskTitle)` because the pointer dies with the copy. Fix by switching to pointer receivers OR by using the heap-stable `formData` pattern for subtasks.

**Files:**
- Modify: `internal/app/model.go`

**Step 1: Remove `subtaskTitle` field and use `formData` for subtask form**

Replace lines 50, 186-191, and 272-279:

```go
// In Model struct, DELETE this line:
// subtaskTitle string

// In the Subtask key handler (around line 186-193), replace with:
case key.Matches(msg, keys.Subtask):
    if m.selectedTask() != nil {
        m.formData = &ui.TaskFormData{}
        m.form = ui.SubtaskForm(&m.formData.Title)
        m.mode = modeSubtask
        return m, m.form.Init()
    }
    return m, nil

// In updateFormMsg subtask completion (around line 272-283), replace with:
if m.mode == modeSubtask {
    selected := m.selectedTask()
    title := ""
    if m.formData != nil {
        title = m.formData.Title
    }
    if selected != nil && title != "" {
        if err := m.store.AddSubtask(selected.ID, title); err != nil {
            m.err = err
            return m, nil
        }
    }
    m.mode = modeNormal
    m.form = nil
    m.formData = nil
    m.reload()
    return m, nil
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Clean build, no errors.

**Step 3: Commit**

```bash
git add internal/app/model.go
git commit -m "fix: use heap-stable formData for subtask form pointer"
```

---

### Task 2: Fix layout calculations

Three issues: (a) hardcoded `headerHeight = 3` but actual is 2, (b) panel sizing ignores padding, (c) `updateDetail` width inconsistent with viewport.

**Files:**
- Modify: `internal/app/model.go`

**Step 1: Fix `resizeComponents` to use dynamic header height and account for padding**

The styles have these frame sizes:
- `listPanelFocusedStyle`: border=2w,2h + padding(0,1)=2w,0h → total frame: 4w, 2h
- `detailPanelStyle`: border=2w,2h + padding(1,2)=4w,2h → total frame: 6w, 4h

Replace `resizeComponents` entirely:

```go
func (m *Model) resizeComponents() {
	header := ui.RenderTabs(m.activeTab, 0, 0, 0, m.width)
	headerHeight := lipgloss.Height(header)
	statusBarHeight := 1
	contentHeight := m.height - headerHeight - statusBarHeight

	listWidth := m.width * 2 / 5
	detailWidth := m.width - listWidth

	// list panel: border(2) + horizontal padding(2) = 4 wide, border(2) high
	m.list.SetSize(listWidth-4, contentHeight-2)
	// detail panel: border(2) + horizontal padding(4) = 6 wide, border(2) + vertical padding(2) = 4 high
	m.viewport.Width = detailWidth - 6
	m.viewport.Height = contentHeight - 4
	m.help.Width = m.width
}
```

**Step 2: Fix `updateDetail` to use viewport width**

Replace the `updateDetail` function:

```go
func (m *Model) updateDetail() {
	selected := m.selectedTask()
	content := ui.RenderDetail(selected, m.viewport.Width)
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}
```

**Step 3: Add `updateDetail()` call after WindowSizeMsg**

In the `tea.WindowSizeMsg` handler (around line 114-119), add `m.updateDetail()`:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.ready = true
    m.resizeComponents()
    m.updateDetail()
    return m, nil
```

**Step 4: Build and verify**

Run: `go build ./...`
Expected: Clean build.

**Step 5: Commit**

```bash
git add internal/app/model.go
git commit -m "fix: correct layout calculations for header height, padding, and viewport"
```

---

### Task 3: Fix form lifecycle — handle StateAborted

When huh internally handles Esc (e.g., in a select field), the form reaches `StateAborted` instead of propagating Esc as a KeyMsg. The current code only checks `StateCompleted`, leaving the mode stuck.

**Files:**
- Modify: `internal/app/model.go`

**Step 1: Add StateAborted check in updateFormMsg**

After the `StateCompleted` check block (around line 287), add:

```go
if f, ok := form.(*huh.Form); ok {
    m.form = f
    if m.form.State == huh.StateCompleted {
        // ... existing completion logic ...
    }
    if m.form.State == huh.StateAborted {
        m.mode = modeNormal
        m.form = nil
        m.formData = nil
        return m, nil
    }
}
```

**Step 2: Build and verify**

Run: `go build ./...`

**Step 3: Commit**

```bash
git add internal/app/model.go
git commit -m "fix: handle huh form StateAborted to prevent stuck mode"
```

---

### Task 4: Fix state management — preserve selection after reload

`list.SetItems()` resets the cursor to index 0. After any mutation (add, delete, toggle status, add subtask), the user loses their place in the list.

**Files:**
- Modify: `internal/app/model.go`

**Step 1: Preserve and restore selection in `reload()`**

Replace `reload()`:

```go
func (m *Model) reload() {
	// Remember current selection.
	var selectedID int64
	if sel := m.selectedTask(); sel != nil {
		selectedID = sel.ID
	}

	tasks, err := m.store.List()
	if err != nil {
		m.err = err
		return
	}
	m.tasks = tasks
	m.sortTasks()

	// Restore selection.
	if selectedID != 0 {
		for i, item := range m.list.Items() {
			if t, ok := item.(task.Task); ok && t.ID == selectedID {
				m.list.Select(i)
				break
			}
		}
		m.updateDetail()
	}
}
```

Note: `sortTasks()` already calls `refreshList()` and `updateDetail()`, so remove the redundant calls that were after `sortTasks()`.

**Step 2: Build and verify**

Run: `go build ./...`

**Step 3: Commit**

```bash
git add internal/app/model.go
git commit -m "fix: preserve list selection across reload"
```

---

### Task 5: Fix confirm delete dialog and error handling

Two issues: (a) confirm delete replaces entire view instead of overlaying, (b) errors permanently kill the UI.

**Files:**
- Modify: `internal/app/model.go`
- Modify: `internal/ui/views.go`

**Step 1: Change confirm delete to overlay pattern**

In `View()`, replace the `modeConfirmDelete` case (around line 446-452):

```go
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
```

In `internal/ui/views.go`, rename `RenderConfirmDialog` to `RenderConfirmDialogBox` and have it return just the styled box (not the full placed view):

```go
// RenderConfirmDialogBox renders the dialog box content without placement.
func RenderConfirmDialogBox(title, message string) string {
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(white).
		Render(title) + "\n\n" +
		lipgloss.NewStyle().Foreground(gray).Render(message) + "\n\n" +
		lipgloss.NewStyle().Foreground(gray).Render("[y] confirm  [n/esc] cancel")

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(red).
		Padding(1, 2).
		Width(50).
		Render(content)
}
```

Remove the old `RenderConfirmDialog` function.

**Step 2: Make errors dismissable**

Replace the error display in `View()` (around line 368-370):

```go
// Remove this block:
// if m.err != nil {
//     return fmt.Sprintf("Error: %v", m.err)
// }
```

Instead, show the error in the status bar. Add a `statusMsg` field to Model:

```go
type Model struct {
    // ... existing fields ...
    statusMsg string
}
```

Add a `clearStatusMsg` message type and handler:

```go
type clearStatusMsg struct{}

// In Update(), add a case before tea.KeyMsg:
case clearStatusMsg:
    m.statusMsg = ""
    return m, nil
```

Add a helper method:

```go
func (m *Model) setStatus(msg string) tea.Cmd {
    m.statusMsg = msg
    return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
        return clearStatusMsg{}
    })
}
```

When `m.err` is set, convert it to a status message and clear the error:

```go
// Add this helper:
func (m *Model) setError(err error) tea.Cmd {
    m.statusMsg = "Error: " + err.Error()
    m.err = nil
    return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
        return clearStatusMsg{}
    })
}
```

Replace all `m.err = err` with `return m, m.setError(err)` throughout the file. This requires adjusting the control flow in several places (status toggle, subtask toggle, delete, create, update).

**Step 3: Pass statusMsg to RenderStatusBar**

Update `RenderStatusBar` signature in `internal/ui/views.go`:

```go
func RenderStatusBar(total, done, inProgress int, width int, statusMsg string) string {
    keyStyle := lipgloss.NewStyle().Foreground(cyan)
    dimStyle := lipgloss.NewStyle().Foreground(gray)

    var left string
    if statusMsg != "" {
        color := green
        if strings.HasPrefix(statusMsg, "Error:") {
            color = red
        }
        left = lipgloss.NewStyle().Foreground(color).Bold(true).Render(" " + statusMsg)
    } else {
        left = dimStyle.Render(fmt.Sprintf(" %d tasks | %d done | %d active", total, done, inProgress))
    }
    // ... rest unchanged
}
```

Update the call in `View()`:

```go
statusBar := ui.RenderStatusBar(allCount, doneCount, inProgress, m.width, m.statusMsg)
```

**Step 4: Add status messages after successful actions**

After each successful store mutation, call `m.setStatus()`:

```go
// After Create: return m, m.setStatus("Task created")
// After Update (edit): return m, m.setStatus("Task updated")
// After Delete: return m, m.setStatus("Task deleted")
// After Status toggle: return m, m.setStatus("Status: " + selected.Status.String())
// After AddSubtask: return m, m.setStatus("Subtask added")
// After ToggleSubtask: return m, m.setStatus("Subtask toggled")
```

This requires restructuring `submitTaskForm()` to return a `tea.Cmd` instead of being void.

**Step 5: Build and verify**

Run: `go build ./...`

**Step 6: Commit**

```bash
git add internal/app/model.go internal/ui/views.go
git commit -m "fix: overlay confirm dialog, inline errors, add action feedback"
```

---

### Task 6: Fix search/filter lifecycle

Two issues: (a) no way to exit filter mode with Escape, (b) Tab fires during filter input.

**Files:**
- Modify: `internal/app/model.go`

**Step 1: Add Escape handler in normal mode to dismiss filter**

In the normal mode `case tea.KeyMsg:` block, add a check at the top (before the switch):

```go
case tea.KeyMsg:
    if m.mode == modeConfirmDelete {
        return m.updateConfirmDelete(msg)
    }
    if m.mode == modeHelp {
        // ... existing help handler
    }

    // Dismiss active filter with Escape.
    if key.Matches(msg, keys.Escape) {
        if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
            m.list.ResetFilter()
            m.list.SetShowFilter(false)
            return m, nil
        }
    }

    // Normal mode keybindings.
    switch {
    // ...
```

**Step 2: Guard Tab during filtering**

Update the Tab handler:

```go
case key.Matches(msg, keys.Tab):
    if m.list.FilterState() == list.Filtering {
        // Let the list handle Tab during active filtering
        break
    }
    m.activeTab = (m.activeTab + 1) % 3
    m.refreshList()
    m.updateDetail()
    return m, nil
```

Note: when Tab breaks out here instead of returning, it falls through to `m.list.Update(msg)` below, which passes Tab to the list for filter navigation.

**Step 3: Build and verify**

Run: `go build ./...`

**Step 4: Commit**

```bash
git add internal/app/model.go
git commit -m "fix: add Escape to dismiss filter, guard Tab during filtering"
```

---

### Task 7: UX polish — empty states, form help, help overlay

**Files:**
- Modify: `internal/app/model.go`
- Modify: `internal/ui/form.go`
- Modify: `internal/ui/views.go`

**Step 1: Add empty state messages in View()**

In `View()`, after building the list view but before creating `listPanel`, check if the list is empty:

```go
var listContent string
if len(m.list.Items()) == 0 {
    var emptyText string
    switch m.activeTab {
    case 1:
        emptyText = "No active tasks\n\nAll tasks are completed!"
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

listPanel := listPanelFocusedStyle.
    Width(listWidth - 2).
    Height(contentHeight - 2).
    Render(listContent)
```

**Step 2: Enable form help hints**

In `internal/ui/form.go`, change `.WithShowHelp(false)` to `.WithShowHelp(true)` on all three form constructors (lines 56, 94, 106).

**Step 3: Add form navigation section to help overlay**

In `renderHelpOverlay` in `model.go`, add form navigation entries after the existing ones:

```go
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
    {"Esc", "Clear filter"},
    {"Tab", "Switch tab"},
    {"F1/F2/F3", "Sort date/due/prio"},
    {"?", "Toggle this help"},
    {"q", "Quit"},
    {"", ""},
    {"", "In Forms:"},
    {"Tab/Shift+Tab", "Next / prev field"},
    {"Enter", "Submit form"},
    {"Esc", "Cancel"},
}
```

And update the rendering loop to handle section headers:

```go
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
```

Widen the help box to accommodate the wider key column:

```go
return lipgloss.NewStyle().
    Border(lipgloss.DoubleBorder()).
    BorderForeground(cyan).
    Padding(1, 3).
    Width(50).  // was 45
    Render(content)
```

**Step 4: Improve priority labels**

In `internal/task/task.go`, replace the `Label()` method:

```go
func (p Priority) Label() string {
    switch p {
    case Medium:
        return "MED"
    case High:
        return "HIGH"
    case Urgent:
        return "URG!"
    default:
        return "LOW"
    }
}
```

**Step 5: Build and verify**

Run: `go build ./...`

**Step 6: Commit**

```bash
git add internal/app/model.go internal/ui/form.go internal/ui/views.go internal/task/task.go
git commit -m "feat: add empty states, form help, improved help overlay, readable priority labels"
```

---

### Task 8: Final integration verification

**Files:** All modified files

**Step 1: Full build**

Run: `go build ./...`
Expected: Clean build.

**Step 2: Vet**

Run: `go vet ./...`
Expected: No issues.

**Step 3: Manual smoke test**

Run: `go run ./cmd/todo/`

Test checklist:
- [ ] Tabs visible at top of screen
- [ ] Layout fills terminal correctly (no overflow, no blank rows)
- [ ] Press `a` → form appears with help hints, Tab navigates fields, Enter submits, status message shows "Task created"
- [ ] Press `t` on a task → subtask form works, subtask appears in detail panel after submit
- [ ] Press `s` → status cycles, status message confirms
- [ ] Press `d` → confirm dialog overlays (not replaces) the view
- [ ] Press `/` → filter activates, Esc dismisses it
- [ ] Press `?` → help shows with form navigation section
- [ ] Empty list shows helpful message
- [ ] Switch tabs → cursor preserved on return
- [ ] Narrow terminal → status bar drops right side gracefully

**Step 4: Commit any remaining fixes**

```bash
git add -A
git commit -m "chore: final integration fixes"
```
