package task

import "testing"

func TestHasCycle_NoCycle(t *testing.T) {
	// A -> B -> C, adding D as blocker of A should be fine.
	blockers := map[int64][]int64{
		1: {2},
		2: {3},
		3: {},
		4: {},
	}
	getBlockers := func(id int64) []int64 { return blockers[id] }
	if HasCycle(1, []int64{4}, getBlockers) {
		t.Error("expected no cycle when adding independent task as blocker")
	}
}

func TestHasCycle_DirectCycle(t *testing.T) {
	// A -> B, adding A as blocker of B creates B -> A -> B cycle.
	blockers := map[int64][]int64{
		1: {2},
		2: {},
	}
	getBlockers := func(id int64) []int64 { return blockers[id] }
	if !HasCycle(2, []int64{1}, getBlockers) {
		t.Error("expected cycle: B blocked by A, which is blocked by B")
	}
}

func TestHasCycle_TransitiveCycle(t *testing.T) {
	// A -> B -> C, adding A as blocker of C creates C -> A -> B -> C.
	blockers := map[int64][]int64{
		1: {2},
		2: {3},
		3: {},
	}
	getBlockers := func(id int64) []int64 { return blockers[id] }
	if !HasCycle(3, []int64{1}, getBlockers) {
		t.Error("expected transitive cycle")
	}
}

func TestHasCycle_SelfCycle(t *testing.T) {
	blockers := map[int64][]int64{
		1: {},
	}
	getBlockers := func(id int64) []int64 { return blockers[id] }
	if !HasCycle(1, []int64{1}, getBlockers) {
		t.Error("expected self-cycle")
	}
}

func TestHasCycle_MultipleBlockers(t *testing.T) {
	// A -> B, A -> C. Adding A as blocker of B should detect cycle through B -> A -> B.
	blockers := map[int64][]int64{
		1: {2, 3},
		2: {},
		3: {},
	}
	getBlockers := func(id int64) []int64 { return blockers[id] }
	if !HasCycle(2, []int64{1}, getBlockers) {
		t.Error("expected cycle via one of multiple blockers")
	}
}

func TestHasCycle_EmptyBlockers(t *testing.T) {
	getBlockers := func(id int64) []int64 { return nil }
	if HasCycle(1, nil, getBlockers) {
		t.Error("empty blockerIDs should never create a cycle")
	}
	if HasCycle(1, []int64{}, getBlockers) {
		t.Error("empty blockerIDs should never create a cycle")
	}
}

func TestIsBlocked_AllDone(t *testing.T) {
	getStatus := func(id int64) Status { return Done }
	if IsBlocked([]int64{1, 2, 3}, getStatus) {
		t.Error("all blockers are Done, should not be blocked")
	}
}

func TestIsBlocked_OnePending(t *testing.T) {
	statuses := map[int64]Status{
		1: Done,
		2: Pending,
		3: Done,
	}
	getStatus := func(id int64) Status { return statuses[id] }
	if !IsBlocked([]int64{1, 2, 3}, getStatus) {
		t.Error("one blocker is Pending, should be blocked")
	}
}

func TestIsBlocked_OneInProgress(t *testing.T) {
	statuses := map[int64]Status{
		1: Done,
		2: InProgress,
	}
	getStatus := func(id int64) Status { return statuses[id] }
	if !IsBlocked([]int64{1, 2}, getStatus) {
		t.Error("one blocker is InProgress, should be blocked")
	}
}

func TestIsBlocked_Empty(t *testing.T) {
	getStatus := func(id int64) Status { return Pending }
	if IsBlocked(nil, getStatus) {
		t.Error("no blockers means not blocked")
	}
	if IsBlocked([]int64{}, getStatus) {
		t.Error("empty blockers means not blocked")
	}
}
