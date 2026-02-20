package ui

import (
	"fmt"
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
					huh.NewOption("Low", task.Low),
					huh.NewOption("Medium", task.Medium).Selected(true),
					huh.NewOption("High", task.High),
					huh.NewOption("Urgent", task.Urgent),
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
	).WithTheme(huh.ThemeDracula()).WithShowHelp(false)
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
	).WithTheme(huh.ThemeDracula()).WithShowHelp(false)
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
	).WithTheme(huh.ThemeDracula()).WithShowHelp(false)
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
