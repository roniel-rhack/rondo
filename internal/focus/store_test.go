package focus

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	// Enable foreign keys for consistency with production.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewStore(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestCreateAndListByTask(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	sess := &Session{
		TaskID:    42,
		Duration:  DefaultDuration,
		StartedAt: now,
	}

	if err := store.Create(sess); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sess.ID == 0 {
		t.Error("expected session ID to be set after Create")
	}

	sessions, err := store.ListByTask(42)
	if err != nil {
		t.Fatalf("ListByTask: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	got := sessions[0]
	if got.ID != sess.ID {
		t.Errorf("ID = %d, want %d", got.ID, sess.ID)
	}
	if got.TaskID != 42 {
		t.Errorf("TaskID = %d, want 42", got.TaskID)
	}
	if got.Duration != DefaultDuration {
		t.Errorf("Duration = %v, want %v", got.Duration, DefaultDuration)
	}
	if got.IsCompleted() {
		t.Error("expected incomplete session")
	}
}

func TestComplete(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	sess := &Session{
		TaskID:    1,
		Duration:  DefaultDuration,
		StartedAt: time.Now().Truncate(time.Second),
	}
	if err := store.Create(sess); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := store.Complete(sess.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	sessions, err := store.ListByTask(1)
	if err != nil {
		t.Fatalf("ListByTask: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if !sessions[0].IsCompleted() {
		t.Error("expected session to be completed")
	}
}

func TestCompleteNotFound(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if err := store.Complete(999); err == nil {
		t.Error("expected error completing non-existent session")
	}
}

func TestListByTaskEmpty(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	sessions, err := store.ListByTask(999)
	if err != nil {
		t.Fatalf("ListByTask: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestListByTaskOrdering(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		sess := &Session{
			TaskID:    5,
			Duration:  DefaultDuration,
			StartedAt: base.Add(time.Duration(i) * time.Hour),
		}
		if err := store.Create(sess); err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
	}

	sessions, err := store.ListByTask(5)
	if err != nil {
		t.Fatalf("ListByTask: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}
	// Should be ordered DESC by started_at.
	if !sessions[0].StartedAt.After(sessions[1].StartedAt) {
		t.Error("sessions not ordered by started_at DESC")
	}
	if !sessions[1].StartedAt.After(sessions[2].StartedAt) {
		t.Error("sessions not ordered by started_at DESC")
	}
}

func TestTodayCount(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	count, err := store.TodayCount()
	if err != nil {
		t.Fatalf("TodayCount: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Create and complete a session today.
	now := time.Now().Truncate(time.Second)
	sess := &Session{
		TaskID:    1,
		Duration:  DefaultDuration,
		StartedAt: now.Add(-30 * time.Minute),
	}
	if err := store.Create(sess); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.Complete(sess.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	count, err = store.TodayCount()
	if err != nil {
		t.Fatalf("TodayCount: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestCompletionsByDay(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	now := time.Now().Truncate(time.Second)

	// Create 2 completed sessions today.
	for i := 0; i < 2; i++ {
		sess := &Session{
			TaskID:    1,
			Duration:  DefaultDuration,
			StartedAt: now.Add(time.Duration(-i) * time.Hour),
		}
		if err := store.Create(sess); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if err := store.Complete(sess.ID); err != nil {
			t.Fatalf("Complete: %v", err)
		}
	}

	// Create 1 incomplete session (should not be counted).
	incomplete := &Session{
		TaskID:    1,
		Duration:  DefaultDuration,
		StartedAt: now,
	}
	if err := store.Create(incomplete); err != nil {
		t.Fatalf("Create: %v", err)
	}

	result, err := store.CompletionsByDay(7)
	if err != nil {
		t.Fatalf("CompletionsByDay: %v", err)
	}

	today := now.Format(time.DateOnly)
	if result[today] != 2 {
		t.Errorf("expected 2 completions today, got %d", result[today])
	}
}

func TestCreateWithZeroTaskID(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	sess := &Session{
		TaskID:    0,
		Duration:  DefaultDuration,
		StartedAt: time.Now().Truncate(time.Second),
	}
	if err := store.Create(sess); err != nil {
		t.Fatalf("Create with zero TaskID: %v", err)
	}

	sessions, err := store.ListByTask(0)
	if err != nil {
		t.Fatalf("ListByTask(0): %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].TaskID != 0 {
		t.Errorf("TaskID = %d, want 0", sessions[0].TaskID)
	}
}
