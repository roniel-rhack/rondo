package task

// HasCycle performs DFS cycle detection. It returns true if adding blockerIDs
// as blockers of taskID would create a dependency cycle. The getBlockers
// function returns the existing blocker IDs for any given task ID.
func HasCycle(taskID int64, blockerIDs []int64, getBlockers func(int64) []int64) bool {
	// We need to check: if taskID depends on each blockerID, does any
	// blockerID (transitively) already depend on taskID?
	// That is equivalent to: can we reach taskID by following the blocker
	// graph starting from any of the proposed blockerIDs?
	visited := make(map[int64]bool)

	var dfs func(id int64) bool
	dfs = func(id int64) bool {
		if id == taskID {
			return true
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		for _, dep := range getBlockers(id) {
			if dfs(dep) {
				return true
			}
		}
		return false
	}

	for _, bid := range blockerIDs {
		// Reset visited for each starting blocker to ensure correct traversal.
		visited = make(map[int64]bool)
		if dfs(bid) {
			return true
		}
	}
	return false
}

// IsBlocked returns true if any of the blocker task IDs has a status that is
// not Done.
func IsBlocked(blockedByIDs []int64, getStatus func(int64) Status) bool {
	for _, id := range blockedByIDs {
		if getStatus(id) != Done {
			return true
		}
	}
	return false
}
