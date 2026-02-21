package cli

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roniel/todo-app/internal/journal"
	"github.com/roniel/todo-app/internal/task"

	_ "modernc.org/sqlite"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestStores(t *testing.T) (*task.Store, *journal.Store) {
	t.Helper()
	db := openTestDB(t)
	ts, err := task.NewStore(db)
	if err != nil {
		t.Fatalf("task.NewStore: %v", err)
	}
	js, err := journal.NewStore(db)
	if err != nil {
		t.Fatalf("journal.NewStore: %v", err)
	}
	return ts, js
}

// captureStdout runs fn while capturing everything written to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	buf, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(buf)
}

// ---------------------------------------------------------------------------
// add command
// ---------------------------------------------------------------------------

func TestIntegration_Add_Basic(t *testing.T) {
	ts, _ := newTestStores(t)

	if err := cmdAdd([]string{"Buy milk"}, ts); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}

	tasks, err := ts.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	got := tasks[0]
	if got.Title != "Buy milk" {
		t.Errorf("Title = %q, want %q", got.Title, "Buy milk")
	}
	if got.Priority != task.Low {
		t.Errorf("Priority = %v, want Low", got.Priority)
	}
	if got.Status != task.Pending {
		t.Errorf("Status = %v, want Pending", got.Status)
	}
}

func TestIntegration_Add_AllFlags(t *testing.T) {
	ts, _ := newTestStores(t)

	args := []string{"--priority", "high", "--due", "2026-03-15", "--tags", "home,shopping", "Big task"}
	if err := cmdAdd(args, ts); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}

	tasks, err := ts.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	got := tasks[0]
	if got.Title != "Big task" {
		t.Errorf("Title = %q, want %q", got.Title, "Big task")
	}
	if got.Priority != task.High {
		t.Errorf("Priority = %v, want High", got.Priority)
	}
	if got.DueDate == nil {
		t.Fatal("expected DueDate to be set")
	}
	if got.DueDate.Format("2006-01-02") != "2026-03-15" {
		t.Errorf("DueDate = %s, want 2026-03-15", got.DueDate.Format("2006-01-02"))
	}
	if len(got.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(got.Tags))
	}
	if got.Tags[0] != "home" || got.Tags[1] != "shopping" {
		t.Errorf("Tags = %v, want [home shopping]", got.Tags)
	}
}

func TestIntegration_Add_Multiple(t *testing.T) {
	ts, _ := newTestStores(t)

	for _, title := range []string{"Task 1", "Task 2", "Task 3"} {
		if err := cmdAdd([]string{title}, ts); err != nil {
			t.Fatalf("cmdAdd(%q): %v", title, err)
		}
	}

	tasks, err := ts.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

// ---------------------------------------------------------------------------
// done command
// ---------------------------------------------------------------------------

func TestIntegration_Done(t *testing.T) {
	ts, _ := newTestStores(t)

	if err := cmdAdd([]string{"Finish report"}, ts); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}

	tasks, err := ts.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	id := tasks[0].ID

	if err := cmdDone([]string{"1"}, ts); err != nil {
		t.Fatalf("cmdDone: %v", err)
	}

	got, err := ts.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Status != task.Done {
		t.Errorf("Status = %v, want Done", got.Status)
	}
}

func TestIntegration_Done_NotFound(t *testing.T) {
	ts, _ := newTestStores(t)

	err := cmdDone([]string{"999"}, ts)
	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "not found")
	}
}

// ---------------------------------------------------------------------------
// list command
// ---------------------------------------------------------------------------

func TestIntegration_List_Table(t *testing.T) {
	ts, _ := newTestStores(t)

	cmdAdd([]string{"Alpha"}, ts)
	cmdAdd([]string{"Beta"}, ts)

	out := captureStdout(t, func() {
		if err := cmdList(nil, ts); err != nil {
			t.Fatalf("cmdList: %v", err)
		}
	})

	if !strings.Contains(out, "Alpha") {
		t.Errorf("output missing %q:\n%s", "Alpha", out)
	}
	if !strings.Contains(out, "Beta") {
		t.Errorf("output missing %q:\n%s", "Beta", out)
	}
	if !strings.Contains(out, "TITLE") {
		t.Errorf("output missing header %q:\n%s", "TITLE", out)
	}
}

func TestIntegration_List_JSON(t *testing.T) {
	ts, _ := newTestStores(t)

	cmdAdd([]string{"JSON task"}, ts)

	out := captureStdout(t, func() {
		if err := cmdList([]string{"--format", "json"}, ts); err != nil {
			t.Fatalf("cmdList json: %v", err)
		}
	})

	var result []json.RawMessage
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "JSON task") {
		t.Errorf("output missing %q:\n%s", "JSON task", out)
	}
}

func TestIntegration_List_FilterDone(t *testing.T) {
	ts, _ := newTestStores(t)

	cmdAdd([]string{"Stay pending"}, ts)
	cmdAdd([]string{"Mark done"}, ts)
	cmdDone([]string{"2"}, ts)

	out := captureStdout(t, func() {
		if err := cmdList([]string{"--status", "done"}, ts); err != nil {
			t.Fatalf("cmdList --status done: %v", err)
		}
	})

	if !strings.Contains(out, "Mark done") {
		t.Errorf("output missing done task:\n%s", out)
	}
	if strings.Contains(out, "Stay pending") {
		t.Errorf("output should not contain pending task:\n%s", out)
	}
}

func TestIntegration_List_FilterPending(t *testing.T) {
	ts, _ := newTestStores(t)

	cmdAdd([]string{"Pending one"}, ts)
	cmdAdd([]string{"Done one"}, ts)
	cmdDone([]string{"2"}, ts)

	out := captureStdout(t, func() {
		if err := cmdList([]string{"--status", "pending"}, ts); err != nil {
			t.Fatalf("cmdList --status pending: %v", err)
		}
	})

	if !strings.Contains(out, "Pending one") {
		t.Errorf("output missing pending task:\n%s", out)
	}
	if strings.Contains(out, "Done one") {
		t.Errorf("output should not contain done task:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// journal command
// ---------------------------------------------------------------------------

func TestIntegration_JournalAdd(t *testing.T) {
	_, js := newTestStores(t)

	if err := cmdJournalAdd([]string{"Great day"}, js); err != nil {
		t.Fatalf("cmdJournalAdd: %v", err)
	}

	notes, err := js.ListNotes(false)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	entries := notes[0].Entries
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Great day" {
		t.Errorf("Body = %q, want %q", entries[0].Body, "Great day")
	}
}

func TestIntegration_JournalAdd_MultipleWords(t *testing.T) {
	_, js := newTestStores(t)

	if err := cmdJournalAdd([]string{"Hello", "world"}, js); err != nil {
		t.Fatalf("cmdJournalAdd: %v", err)
	}

	notes, err := js.ListNotes(false)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Entries[0].Body != "Hello world" {
		t.Errorf("Body = %q, want %q", notes[0].Entries[0].Body, "Hello world")
	}
}

// ---------------------------------------------------------------------------
// export command
// ---------------------------------------------------------------------------

func TestIntegration_Export_Markdown(t *testing.T) {
	ts, js := newTestStores(t)

	cmdAdd([]string{"Export me"}, ts)

	out := captureStdout(t, func() {
		if err := cmdExport([]string{"--format", "md"}, ts, js); err != nil {
			t.Fatalf("cmdExport md: %v", err)
		}
	})

	if !strings.Contains(out, "# Tasks") {
		t.Errorf("output missing %q:\n%s", "# Tasks", out)
	}
	if !strings.Contains(out, "Export me") {
		t.Errorf("output missing task title:\n%s", out)
	}
}

func TestIntegration_Export_JSON(t *testing.T) {
	ts, js := newTestStores(t)

	cmdAdd([]string{"JSON export"}, ts)

	out := captureStdout(t, func() {
		if err := cmdExport([]string{"--format", "json"}, ts, js); err != nil {
			t.Fatalf("cmdExport json: %v", err)
		}
	})

	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "JSON export") {
		t.Errorf("output missing task title:\n%s", out)
	}
}

func TestIntegration_Export_ToFile(t *testing.T) {
	ts, js := newTestStores(t)

	cmdAdd([]string{"File export"}, ts)

	tmpFile := filepath.Join(t.TempDir(), "export.md")
	if err := cmdExport([]string{"--format", "md", "--output", tmpFile}, ts, js); err != nil {
		t.Fatalf("cmdExport to file: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read export file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Tasks") {
		t.Errorf("file missing %q:\n%s", "# Tasks", content)
	}
	if !strings.Contains(content, "File export") {
		t.Errorf("file missing task title:\n%s", content)
	}
}

func TestIntegration_Export_WithJournal(t *testing.T) {
	ts, js := newTestStores(t)

	cmdAdd([]string{"My task"}, ts)
	cmdJournalAdd([]string{"My entry"}, js)

	out := captureStdout(t, func() {
		if err := cmdExport([]string{"--format", "md", "--journal"}, ts, js); err != nil {
			t.Fatalf("cmdExport with journal: %v", err)
		}
	})

	if !strings.Contains(out, "# Tasks") {
		t.Errorf("output missing %q:\n%s", "# Tasks", out)
	}
	if !strings.Contains(out, "# Journal") {
		t.Errorf("output missing %q:\n%s", "# Journal", out)
	}
}

// ---------------------------------------------------------------------------
// dispatch (Run)
// ---------------------------------------------------------------------------

func TestIntegration_Run_Dispatch(t *testing.T) {
	ts, js := newTestStores(t)

	if err := Run([]string{"add", "My task"}, ts, js); err != nil {
		t.Fatalf("Run add: %v", err)
	}

	tasks, err := ts.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "My task" {
		t.Errorf("Title = %q, want %q", tasks[0].Title, "My task")
	}
}
