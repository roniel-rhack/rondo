package task

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func createTestTask(t *testing.T, store *Store, title string) *Task {
	t.Helper()
	task := &Task{Title: title}
	if err := store.Create(task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task
}

func TestAddNote(t *testing.T) {
	store := newTestStore(t)
	task := createTestTask(t, store, "test task")

	if err := store.AddNote(task.ID, "first note"); err != nil {
		t.Fatalf("AddNote: %v", err)
	}

	notes, err := store.ListNotes(task.ID)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Body != "first note" {
		t.Errorf("expected body %q, got %q", "first note", notes[0].Body)
	}
	if notes[0].TaskID != task.ID {
		t.Errorf("expected TaskID %d, got %d", task.ID, notes[0].TaskID)
	}
}

func TestUpdateNote(t *testing.T) {
	store := newTestStore(t)
	task := createTestTask(t, store, "test task")
	store.AddNote(task.ID, "original")

	notes, _ := store.ListNotes(task.ID)
	if err := store.UpdateNote(notes[0].ID, "updated"); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}

	notes, _ = store.ListNotes(task.ID)
	if notes[0].Body != "updated" {
		t.Errorf("expected body %q, got %q", "updated", notes[0].Body)
	}
}

func TestDeleteNote(t *testing.T) {
	store := newTestStore(t)
	task := createTestTask(t, store, "test task")
	store.AddNote(task.ID, "note to delete")

	notes, _ := store.ListNotes(task.ID)
	if err := store.DeleteNote(notes[0].ID); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}

	notes, _ = store.ListNotes(task.ID)
	if len(notes) != 0 {
		t.Errorf("expected 0 notes after delete, got %d", len(notes))
	}
}

func TestRestoreNote(t *testing.T) {
	store := newTestStore(t)
	task := createTestTask(t, store, "test task")
	store.AddNote(task.ID, "restore me")

	notes, _ := store.ListNotes(task.ID)
	original := notes[0]
	store.DeleteNote(original.ID)

	if err := store.RestoreNote(task.ID, original.Body, original.CreatedAt); err != nil {
		t.Fatalf("RestoreNote: %v", err)
	}

	notes, _ = store.ListNotes(task.ID)
	if len(notes) != 1 {
		t.Fatalf("expected 1 note after restore, got %d", len(notes))
	}
	if notes[0].Body != "restore me" {
		t.Errorf("expected body %q, got %q", "restore me", notes[0].Body)
	}
}

func TestListBlocksIDs(t *testing.T) {
	store := newTestStore(t)
	a := createTestTask(t, store, "blocker")
	b := createTestTask(t, store, "blocked1")
	c := createTestTask(t, store, "blocked2")

	// a blocks b and c
	store.SetBlocker(b.ID, a.ID)
	store.SetBlocker(c.ID, a.ID)

	ids, err := store.ListBlocksIDs(a.ID)
	if err != nil {
		t.Fatalf("ListBlocksIDs: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(ids))
	}
}

func TestSetBlockerSelfBlock(t *testing.T) {
	store := newTestStore(t)
	task := createTestTask(t, store, "self block")

	err := store.SetBlocker(task.ID, task.ID)
	if err == nil {
		t.Error("expected error for self-block, got nil")
	}
}

func TestSetBlockersWithSelfBlockGuard(t *testing.T) {
	store := newTestStore(t)
	a := createTestTask(t, store, "task A")
	b := createTestTask(t, store, "task B")

	// SetBlockers should silently skip self-block
	err := store.SetBlockers(a.ID, []int64{a.ID, b.ID})
	if err != nil {
		t.Fatalf("SetBlockers: %v", err)
	}

	ids, _ := store.ListBlockerIDs(a.ID)
	if len(ids) != 1 {
		t.Fatalf("expected 1 blocker (self-block skipped), got %d", len(ids))
	}
	if ids[0] != b.ID {
		t.Errorf("expected blocker %d, got %d", b.ID, ids[0])
	}
}

func TestSetBlocksIDs(t *testing.T) {
	store := newTestStore(t)
	a := createTestTask(t, store, "blocker")
	b := createTestTask(t, store, "blocked1")
	c := createTestTask(t, store, "blocked2")

	if err := store.SetBlocksIDs(a.ID, []int64{b.ID, c.ID}); err != nil {
		t.Fatalf("SetBlocksIDs: %v", err)
	}

	ids, _ := store.ListBlocksIDs(a.ID)
	if len(ids) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(ids))
	}

	// Replace with just c
	if err := store.SetBlocksIDs(a.ID, []int64{c.ID}); err != nil {
		t.Fatalf("SetBlocksIDs replace: %v", err)
	}

	ids, _ = store.ListBlocksIDs(a.ID)
	if len(ids) != 1 {
		t.Fatalf("expected 1 block after replace, got %d", len(ids))
	}
	if ids[0] != c.ID {
		t.Errorf("expected block %d, got %d", c.ID, ids[0])
	}
}

func TestSetBlocksIDsSelfBlock(t *testing.T) {
	store := newTestStore(t)
	a := createTestTask(t, store, "task A")
	b := createTestTask(t, store, "task B")

	// Self-block should be silently skipped
	if err := store.SetBlocksIDs(a.ID, []int64{a.ID, b.ID}); err != nil {
		t.Fatalf("SetBlocksIDs: %v", err)
	}

	ids, _ := store.ListBlocksIDs(a.ID)
	if len(ids) != 1 {
		t.Fatalf("expected 1 block (self-block skipped), got %d", len(ids))
	}
}

func TestMarshalParseMetadata(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]string
		want map[string]string
	}{
		{"nil", nil, nil},
		{"empty", map[string]string{}, nil},
		{"single", map[string]string{"key": "val"}, map[string]string{"key": "val"}},
		{"multi", map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "1", "b": "2"}},
		{"comma in value", map[string]string{"notes": "a,b,c"}, map[string]string{"notes": "a,b,c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := marshalMetadata(tt.in)
			got := parseMetadata(s)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d keys, got %d", len(tt.want), len(got))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q: expected %q, got %q", k, v, got[k])
				}
			}
		})
	}
}

func TestDeleteGuard_ListBlocksIDs(t *testing.T) {
	store := newTestStore(t)
	a := createTestTask(t, store, "blocker")
	b := createTestTask(t, store, "blocked")

	store.SetBlocker(b.ID, a.ID)

	ids, err := store.ListBlocksIDs(a.ID)
	if err != nil {
		t.Fatalf("ListBlocksIDs: %v", err)
	}
	if len(ids) != 1 || ids[0] != b.ID {
		t.Errorf("expected [%d], got %v", b.ID, ids)
	}
}

func TestMetadataRoundTrip(t *testing.T) {
	store := newTestStore(t)
	task := &Task{
		Title:    "meta test",
		Metadata: map[string]string{"notes": "a,b,c", "source": "cli"},
	}
	if err := store.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := store.GetByID(task.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Metadata["notes"] != "a,b,c" {
		t.Errorf("expected notes=a,b,c, got %q", got.Metadata["notes"])
	}
	if got.Metadata["source"] != "cli" {
		t.Errorf("expected source=cli, got %q", got.Metadata["source"])
	}
}
