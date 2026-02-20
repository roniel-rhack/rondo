package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"
)

func dueDate(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return &t
}

func sampleTasks() []task.Task {
	return []task.Task{
		{
			ID:          1,
			Title:       "Write tests",
			Description: "Unit tests for export package",
			Status:      task.InProgress,
			Priority:    task.High,
			DueDate:     dueDate(2026, 3, 1),
			CreatedAt:   time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC),
			Subtasks: []task.Subtask{
				{ID: 1, Title: "Markdown tests", Completed: true},
				{ID: 2, Title: "JSON tests", Completed: false},
			},
			Tags: []string{"dev", "testing"},
		},
		{
			ID:        2,
			Title:     "Deploy app",
			Status:    task.Done,
			Priority:  task.Low,
			CreatedAt: time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		},
	}
}

func sampleNotes() []journal.Note {
	return []journal.Note{
		{
			ID:   1,
			Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.Local),
			Entries: []journal.Entry{
				{
					ID:        1,
					NoteID:    1,
					Body:      "Started Phase 1A implementation",
					CreatedAt: time.Date(2026, 2, 20, 9, 30, 0, 0, time.UTC),
				},
				{
					ID:        2,
					NoteID:    1,
					Body:      "Tests passing",
					CreatedAt: time.Date(2026, 2, 20, 14, 15, 0, 0, time.UTC),
				},
			},
		},
		{
			ID:   2,
			Date: time.Date(2026, 2, 19, 0, 0, 0, 0, time.Local),
			Entries: []journal.Entry{
				{
					ID:        3,
					NoteID:    2,
					Body:      "Planning session",
					CreatedAt: time.Date(2026, 2, 19, 11, 0, 0, 0, time.UTC),
				},
			},
		},
	}
}

// --- Markdown: WriteTasks ---

func TestWriteTasks_Header(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTasks(&buf, sampleTasks()); err != nil {
		t.Fatalf("WriteTasks() error: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "# Tasks\n") {
		t.Error("expected output to start with '# Tasks' header")
	}
}

func TestWriteTasks_CheckboxFormat(t *testing.T) {
	var buf bytes.Buffer
	tasks := sampleTasks()
	if err := WriteTasks(&buf, tasks); err != nil {
		t.Fatalf("WriteTasks() error: %v", err)
	}
	out := buf.String()

	// InProgress task -> unchecked
	if !strings.Contains(out, "- [ ] **Write tests**") {
		t.Error("expected unchecked checkbox for InProgress task")
	}
	// Done task -> checked
	if !strings.Contains(out, "- [x] **Deploy app**") {
		t.Error("expected checked checkbox for Done task")
	}
}

func TestWriteTasks_Metadata(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTasks(&buf, sampleTasks()); err != nil {
		t.Fatalf("WriteTasks() error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "High") {
		t.Error("expected priority 'High' in metadata")
	}
	if !strings.Contains(out, "In Progress") {
		t.Error("expected status 'In Progress' in metadata")
	}
	if !strings.Contains(out, "due 2026-03-01") {
		t.Error("expected due date in metadata")
	}
	if !strings.Contains(out, "tags: dev, testing") {
		t.Error("expected tags in metadata")
	}
}

func TestWriteTasks_Description(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTasks(&buf, sampleTasks()); err != nil {
		t.Fatalf("WriteTasks() error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "> Unit tests for export package") {
		t.Error("expected blockquote description")
	}
}

func TestWriteTasks_Subtasks(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTasks(&buf, sampleTasks()); err != nil {
		t.Fatalf("WriteTasks() error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "  - [x] Markdown tests") {
		t.Error("expected completed subtask with indented checkbox")
	}
	if !strings.Contains(out, "  - [ ] JSON tests") {
		t.Error("expected incomplete subtask with indented checkbox")
	}
}

func TestWriteTasks_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTasks(&buf, nil); err != nil {
		t.Fatalf("WriteTasks() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Tasks") {
		t.Error("expected header even for empty tasks")
	}
	if !strings.Contains(out, "_No tasks._") {
		t.Error("expected empty message for nil tasks")
	}
}

// --- Markdown: WriteNotes ---

func TestWriteNotes_Header(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteNotes(&buf, sampleNotes()); err != nil {
		t.Fatalf("WriteNotes() error: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "# Journal\n") {
		t.Error("expected output to start with '# Journal' header")
	}
}

func TestWriteNotes_DateHeadings(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteNotes(&buf, sampleNotes()); err != nil {
		t.Fatalf("WriteNotes() error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "## 2026-02-20") {
		t.Error("expected date heading for first note")
	}
	if !strings.Contains(out, "## 2026-02-19") {
		t.Error("expected date heading for second note")
	}
}

func TestWriteNotes_EntryFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteNotes(&buf, sampleNotes()); err != nil {
		t.Fatalf("WriteNotes() error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "**09:30**") {
		t.Error("expected timestamp 09:30")
	}
	if !strings.Contains(out, "Started Phase 1A implementation") {
		t.Error("expected first entry body")
	}
	if !strings.Contains(out, "**14:15**") {
		t.Error("expected timestamp 14:15")
	}
	if !strings.Contains(out, "Tests passing") {
		t.Error("expected second entry body")
	}
}

func TestWriteNotes_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteNotes(&buf, nil); err != nil {
		t.Fatalf("WriteNotes() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Journal") {
		t.Error("expected header even for empty notes")
	}
	if !strings.Contains(out, "_No journal entries._") {
		t.Error("expected empty message for nil notes")
	}
}

// --- JSON ---

func TestWriteJSON_TopLevelKeys(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, sampleTasks(), sampleNotes()); err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := result["tasks"]; !ok {
		t.Error("expected 'tasks' key in JSON output")
	}
	if _, ok := result["journal"]; !ok {
		t.Error("expected 'journal' key in JSON output")
	}
}

func TestWriteJSON_TaskFields(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, sampleTasks(), nil); err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	var result struct {
		Tasks []jsonTask `json:"tasks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}

	first := result.Tasks[0]
	if first.ID != 1 {
		t.Errorf("expected ID 1, got %d", first.ID)
	}
	if first.Title != "Write tests" {
		t.Errorf("expected title 'Write tests', got %q", first.Title)
	}
	if first.Status != "In Progress" {
		t.Errorf("expected status 'In Progress', got %q", first.Status)
	}
	if first.Priority != "High" {
		t.Errorf("expected priority 'High', got %q", first.Priority)
	}
	if first.DueDate != "2026-03-01" {
		t.Errorf("expected due date '2026-03-01', got %q", first.DueDate)
	}
	if len(first.Subtasks) != 2 {
		t.Fatalf("expected 2 subtasks, got %d", len(first.Subtasks))
	}
	if !first.Subtasks[0].Completed {
		t.Error("first subtask should be completed")
	}
	if first.Subtasks[1].Completed {
		t.Error("second subtask should not be completed")
	}
	if len(first.Tags) != 2 || first.Tags[0] != "dev" || first.Tags[1] != "testing" {
		t.Errorf("expected tags [dev, testing], got %v", first.Tags)
	}

	// Second task: no due date, no tags, no subtasks.
	second := result.Tasks[1]
	if second.DueDate != "" {
		t.Errorf("expected empty due date, got %q", second.DueDate)
	}
	if len(second.Tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(second.Tags))
	}
	if len(second.Subtasks) != 0 {
		t.Errorf("expected 0 subtasks, got %d", len(second.Subtasks))
	}
}

func TestWriteJSON_NoteFields(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, nil, sampleNotes()); err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	var result struct {
		Journal []jsonNote `json:"journal"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result.Journal) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(result.Journal))
	}

	first := result.Journal[0]
	if first.Date != "2026-02-20" {
		t.Errorf("expected date '2026-02-20', got %q", first.Date)
	}
	if len(first.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(first.Entries))
	}
	if first.Entries[0].Body != "Started Phase 1A implementation" {
		t.Errorf("unexpected entry body: %q", first.Entries[0].Body)
	}
	if first.Entries[0].CreatedAt == "" {
		t.Error("expected non-empty created_at timestamp")
	}
}

func TestWriteJSON_EmptyInputs(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, nil, nil); err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Should still be valid JSON even if empty.
	if _, ok := result["tasks"]; !ok {
		t.Error("expected 'tasks' key even when empty")
	}
}

func TestWriteJSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, sampleTasks(), sampleNotes()); err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	// Verify the output is valid JSON by attempting to unmarshal.
	var raw interface{}
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}
