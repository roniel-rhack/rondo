package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Add           key.Binding
	Edit          key.Binding
	Delete        key.Binding
	Status        key.Binding
	Subtask       key.Binding
	Search        key.Binding
	Tab           key.Binding
	SortDate      key.Binding
	SortDue       key.Binding
	SortPrio      key.Binding
	Hide          key.Binding
	ShowHidden    key.Binding
	Help          key.Binding
	Quit          key.Binding
	Enter         key.Binding
	Escape        key.Binding
	PanelWider    key.Binding
	PanelNarrower key.Binding
	Export        key.Binding
	TimeLog       key.Binding
	Focus         key.Binding
	TagBar        key.Binding
	Undo          key.Binding
	Stats         key.Binding
	Blocker       key.Binding
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
	Hide: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "hide/unhide"),
	),
	ShowHidden: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "show hidden"),
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
	PanelWider: key.NewBinding(
		key.WithKeys(">"),
		key.WithHelp(">", "wider"),
	),
	PanelNarrower: key.NewBinding(
		key.WithKeys("<"),
		key.WithHelp("<", "narrower"),
	),
	Export: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "export"),
	),
	TimeLog: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "log time"),
	),
	Focus: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "focus"),
	),
	TagBar: key.NewBinding(
		key.WithKeys("f4"),
		key.WithHelp("F4", "tags"),
	),
	Undo: key.NewBinding(
		key.WithKeys("ctrl+z"),
		key.WithHelp("ctrl+z", "undo"),
	),
	Stats: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "stats"),
	),
	Blocker: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "blockers"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Edit, k.Delete, k.Status, k.Subtask, k.Search, k.Tab, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Add, k.Edit, k.Delete},
		{k.Status, k.Subtask},
		{k.Search, k.Tab, k.Help},
		{k.SortDate, k.SortDue, k.SortPrio},
		{k.Hide, k.ShowHidden},
		{k.Export, k.Focus, k.Stats},
		{k.PanelWider, k.PanelNarrower, k.Undo},
		{k.Quit},
	}
}
