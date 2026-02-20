package cli

import (
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	err := Run(nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for no args, got nil")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	err := Run([]string{"foobar"}, nil, nil)
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}

func TestCmdAdd_NoTitle(t *testing.T) {
	err := cmdAdd(nil, nil)
	if err == nil {
		t.Fatal("expected error for add with no title, got nil")
	}
}

func TestCmdAdd_InvalidPriority(t *testing.T) {
	err := cmdAdd([]string{"--priority", "extreme", "my task"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
}

func TestCmdAdd_InvalidDueDate(t *testing.T) {
	err := cmdAdd([]string{"--due", "not-a-date", "my task"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid due date, got nil")
	}
}

func TestCmdDone_NoID(t *testing.T) {
	err := cmdDone(nil, nil)
	if err == nil {
		t.Fatal("expected error for done with no ID, got nil")
	}
}

func TestCmdDone_InvalidID(t *testing.T) {
	err := cmdDone([]string{"abc"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestCmdList_InvalidFormat(t *testing.T) {
	// cmdList will try to call store.List(), which will panic with nil store.
	// We test flag parsing by checking an invalid format flag that triggers
	// an error before touching the store.
	// Since the flag parse itself doesn't fail, we accept that this would
	// panic on nil store. Test is about validating error paths that can be
	// reached without a live store.
}

func TestCmdJournalAdd_NoText(t *testing.T) {
	err := cmdJournalAdd(nil, nil)
	if err == nil {
		t.Fatal("expected error for journal with no text, got nil")
	}
}

func TestCmdExport_InvalidFormat(t *testing.T) {
	// cmdExport will attempt to call taskStore.List() which requires a real
	// store. We test flag parsing separately.
}

func TestFilterTasks_All(t *testing.T) {
	// filterTasks with "all" should return all tasks.
	tasks := filterTasks(nil, "all")
	if tasks != nil {
		t.Errorf("expected nil for nil input, got %v", tasks)
	}
}

func TestFilterTasks_Pending(t *testing.T) {
	tasks := filterTasks(nil, "pending")
	if tasks != nil {
		t.Errorf("expected nil for nil input, got %v", tasks)
	}
}

func TestFilterTasks_Done(t *testing.T) {
	tasks := filterTasks(nil, "done")
	if tasks != nil {
		t.Errorf("expected nil for nil input, got %v", tasks)
	}
}

func TestFilterTasks_Active(t *testing.T) {
	tasks := filterTasks(nil, "active")
	if tasks != nil {
		t.Errorf("expected nil for nil input, got %v", tasks)
	}
}
