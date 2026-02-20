package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Add       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Status    key.Binding
	Subtask   key.Binding
	ToggleSub key.Binding
	Search    key.Binding
	Tab       key.Binding
	SortDate  key.Binding
	SortDue   key.Binding
	SortPrio  key.Binding
	Help      key.Binding
	Quit      key.Binding
	Enter     key.Binding
	Escape    key.Binding
}

var keys = keyMap{
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Status: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status"),
	),
	Subtask: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "subtask"),
	),
	ToggleSub: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "toggle sub"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "tab"),
	),
	SortDate: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("F1", "sort date"),
	),
	SortDue: key.NewBinding(
		key.WithKeys("f2"),
		key.WithHelp("F2", "sort due"),
	),
	SortPrio: key.NewBinding(
		key.WithKeys("f3"),
		key.WithHelp("F3", "sort prio"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Edit, k.Delete, k.Status, k.Subtask, k.Search, k.Tab, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Add, k.Edit, k.Delete},
		{k.Status, k.Subtask, k.ToggleSub},
		{k.Search, k.Tab, k.Help},
		{k.SortDate, k.SortDue, k.SortPrio},
		{k.Quit},
	}
}
