package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/roniel/todo-app/internal/task"
)

// FormTheme returns the appropriate Huh form theme for the current terminal background.
func FormTheme() *huh.Theme {
	if IsDark() {
		return huh.ThemeDracula()
	}
	return huh.ThemeBase()
}

// TaskFormData holds the form field values.
type TaskFormData struct {
	Title       string
	Description string
	Priority    task.Priority
	DueDate     string
	Tags        string
	RecurFreq   string
	Metadata    string // "key=value, key2=value2"
	Blocks      string // "1, 2, 3" (task IDs)
}

// NoteFormData holds the note form field value.
type NoteFormData struct {
	Body string
}

// ExportFormData holds the export form field values.
type ExportFormData struct {
	Format         string
	IncludeJournal bool
}

// TimeLogFormData holds the time log form field values.
type TimeLogFormData struct {
	Duration string
	Note     string
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

			huh.NewSelect[string]().
				Title("Recurrence").
				Options(
					huh.NewOption("None", "none").Selected(data.RecurFreq == "" || data.RecurFreq == "none"),
					huh.NewOption("Daily", "daily").Selected(data.RecurFreq == "daily"),
					huh.NewOption("Weekly", "weekly").Selected(data.RecurFreq == "weekly"),
					huh.NewOption("Monthly", "monthly").Selected(data.RecurFreq == "monthly"),
					huh.NewOption("Yearly", "yearly").Selected(data.RecurFreq == "yearly"),
				).
				Value(&data.RecurFreq),

			huh.NewText().
				Title("Metadata").
				Placeholder("key=value (one per line)").
				Value(&data.Metadata).
				Lines(3),

			huh.NewInput().
				Title("Blocks (task IDs)").
				Placeholder("comma separated IDs").
				Value(&data.Blocks).
				Validate(validateOptionalBlocks),
		),
	).WithTheme(FormTheme()).WithShowHelp(true)
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

			huh.NewSelect[string]().
				Title("Recurrence").
				Options(
					huh.NewOption("None", "none").Selected(data.RecurFreq == "" || data.RecurFreq == "none"),
					huh.NewOption("Daily", "daily").Selected(data.RecurFreq == "daily"),
					huh.NewOption("Weekly", "weekly").Selected(data.RecurFreq == "weekly"),
					huh.NewOption("Monthly", "monthly").Selected(data.RecurFreq == "monthly"),
					huh.NewOption("Yearly", "yearly").Selected(data.RecurFreq == "yearly"),
				).
				Value(&data.RecurFreq),

			huh.NewText().
				Title("Metadata").
				Placeholder("key=value (one per line)").
				Value(&data.Metadata).
				Lines(3),

			huh.NewInput().
				Title("Blocks (task IDs)").
				Placeholder("comma separated IDs").
				Value(&data.Blocks).
				Validate(validateOptionalBlocks),
		),
	).WithTheme(FormTheme()).WithShowHelp(true)
}

// NoteForm creates a form for adding or editing a task note.
func NoteForm(data *NoteFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Note").
				Value(&data.Body).
				Lines(3).
				CharLimit(2000).
				Validate(validateNotBlank),
		),
	).WithTheme(FormTheme()).WithShowHelp(true)
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
	).WithTheme(FormTheme()).WithShowHelp(true)
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
	).WithTheme(FormTheme()).WithShowHelp(true)
}

// ExportForm creates a form for choosing export options.
func ExportForm(data *ExportFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Format").
				Options(
					huh.NewOption("Markdown", "md").Selected(data.Format == "" || data.Format == "md"),
					huh.NewOption("JSON", "json").Selected(data.Format == "json"),
				).
				Value(&data.Format),

			huh.NewConfirm().
				Title("Include Journal?").
				Value(&data.IncludeJournal),
		),
	).WithTheme(FormTheme()).WithShowHelp(true)
}

// TimeLogForm creates a form for logging time on a task.
func TimeLogForm(data *TimeLogFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Duration").
				Placeholder("e.g. 1h30m, 45m, 2h").
				Value(&data.Duration).
				Validate(validateDuration),

			huh.NewInput().
				Title("Note (optional)").
				Value(&data.Note),
		),
	).WithTheme(FormTheme()).WithShowHelp(true)
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

func validateDuration(s string) error {
	if s == "" {
		return fmt.Errorf("duration is required")
	}
	_, err := task.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("use format like 1h30m, 45m, 2h")
	}
	return nil
}

func validateOptionalBlocks(s string) error {
	if s == "" {
		return nil
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil || v <= 0 {
			return fmt.Errorf("use comma-separated positive IDs (e.g. 1, 2, 3)")
		}
	}
	return nil
}

func validatePositiveInt(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("enter a whole number")
	}
	if v <= 0 {
		return fmt.Errorf("must be greater than 0")
	}
	return nil
}

// FocusSettingsData holds form field values for focus timer settings.
type FocusSettingsData struct {
	WorkDuration       string
	ShortBreakDuration string
	LongBreakDuration  string
	SessionsPerSet     string
	DailyGoal          string
	AutoStartBreaks    bool
	Sound              bool
}

// FocusSettingsForm creates a Huh form for configuring the focus timer.
func FocusSettingsForm(data *FocusSettingsData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Work Duration (min)").
				Value(&data.WorkDuration).
				Validate(validatePositiveInt),

			huh.NewInput().
				Title("Short Break (min)").
				Value(&data.ShortBreakDuration).
				Validate(validatePositiveInt),

			huh.NewInput().
				Title("Long Break (min)").
				Value(&data.LongBreakDuration).
				Validate(validatePositiveInt),

			huh.NewInput().
				Title("Sessions Per Set").
				Value(&data.SessionsPerSet).
				Validate(validatePositiveInt),

			huh.NewInput().
				Title("Daily Goal").
				Value(&data.DailyGoal).
				Validate(validatePositiveInt),

			huh.NewConfirm().
				Title("Auto-start Breaks?").
				Value(&data.AutoStartBreaks),

			huh.NewConfirm().
				Title("Sound?").
				Value(&data.Sound),
		),
	).WithTheme(FormTheme()).WithShowHelp(true)
}
