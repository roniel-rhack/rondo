package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/roniel/todo-app/internal/task"
)

// TaskFormData holds the form field values.
type TaskFormData struct {
	Title       string
	Description string
	Priority    task.Priority
	DueDate     string
	Tags        string
}

// JournalFormData holds the journal entry form field value.
type JournalFormData struct {
	Body string
}

// NewTaskForm creates a Huh form for adding a new task.
func NewTaskForm(data *TaskFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&data.Title).
				Validate(huh.ValidateNotEmpty()),

			huh.NewText().
				Title("Description").
				Value(&data.Description).
				Lines(3),

			huh.NewSelect[task.Priority]().
				Title("Priority").
				Options(
					huh.NewOption("Low", task.Low).Selected(data.Priority == task.Low),
					huh.NewOption("Medium", task.Medium).Selected(data.Priority == task.Medium),
					huh.NewOption("High", task.High).Selected(data.Priority == task.High),
					huh.NewOption("Urgent", task.Urgent).Selected(data.Priority == task.Urgent),
				).
				Value(&data.Priority),

			huh.NewInput().
				Title("Due Date").
				Placeholder("YYYY-MM-DD").
				Value(&data.DueDate).
				Validate(validateOptionalDate),

			huh.NewInput().
				Title("Tags").
				Placeholder("comma separated").
				Value(&data.Tags),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(true)
}

// EditTaskForm creates a Huh form for editing an existing task.
func EditTaskForm(data *TaskFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&data.Title).
				Validate(huh.ValidateNotEmpty()),

			huh.NewText().
				Title("Description").
				Value(&data.Description).
				Lines(3),

			huh.NewSelect[task.Priority]().
				Title("Priority").
				Options(
					huh.NewOption("Low", task.Low).Selected(data.Priority == task.Low),
					huh.NewOption("Medium", task.Medium).Selected(data.Priority == task.Medium),
					huh.NewOption("High", task.High).Selected(data.Priority == task.High),
					huh.NewOption("Urgent", task.Urgent).Selected(data.Priority == task.Urgent),
				).
				Value(&data.Priority),

			huh.NewInput().
				Title("Due Date").
				Placeholder("YYYY-MM-DD").
				Value(&data.DueDate).
				Validate(validateOptionalDate),

			huh.NewInput().
				Title("Tags").
				Placeholder("comma separated").
				Value(&data.Tags),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(true)
}

// SubtaskForm creates a simple single-field form for adding a subtask.
func SubtaskForm(title *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Subtask").
				Value(title).
				Validate(huh.ValidateNotEmpty()),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(true)
}

// JournalEntryForm creates a form for adding a journal entry.
func JournalEntryForm(body *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Journal Entry").
				Value(body).
				Lines(5).
				CharLimit(2000).
				Validate(validateNotBlank),
		),
	).WithTheme(huh.ThemeDracula()).WithShowHelp(true)
}

func validateNotBlank(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("entry cannot be blank")
	}
	return nil
}

func validateOptionalDate(s string) error {
	if s == "" {
		return nil
	}
	_, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return fmt.Errorf("use YYYY-MM-DD format")
	}
	return nil
}
