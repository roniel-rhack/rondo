package task

import (
	"time"
)

type Status int

const (
	Pending    Status = iota
	InProgress
	Done
)

func (s Status) String() string {
	switch s {
	case InProgress:
		return "In Progress"
	case Done:
		return "Done"
	default:
		return "Pending"
	}
}

func (s Status) Icon() string {
	switch s {
	case InProgress:
		return "◐"
	case Done:
		return "✓"
	default:
		return "○"
	}
}

func (s Status) Next() Status {
	switch s {
	case Pending:
		return InProgress
	case InProgress:
		return Done
	default:
		return Pending
	}
}

type Priority int

const (
	Low Priority = iota
	Medium
	High
	Urgent
)

func (p Priority) String() string {
	switch p {
	case Medium:
		return "Medium"
	case High:
		return "High"
	case Urgent:
		return "Urgent"
	default:
		return "Low"
	}
}

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

type Subtask struct {
	ID        int64
	Title     string
	Completed bool
	Position  int
}

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      Status
	Priority    Priority
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Subtasks    []Subtask
	Tags        []string
}

// FilterValue implements list.Item interface for bubbles list.
func (t Task) FilterValue() string { return t.Title }
