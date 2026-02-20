package task

import "time"

type RecurFreq int

const (
	RecurNone    RecurFreq = iota
	RecurDaily
	RecurWeekly
	RecurMonthly
	RecurYearly
)

func (f RecurFreq) String() string {
	switch f {
	case RecurDaily:
		return "daily"
	case RecurWeekly:
		return "weekly"
	case RecurMonthly:
		return "monthly"
	case RecurYearly:
		return "yearly"
	default:
		return "none"
	}
}

func ParseRecurFreq(s string) RecurFreq {
	switch s {
	case "daily":
		return RecurDaily
	case "weekly":
		return RecurWeekly
	case "monthly":
		return RecurMonthly
	case "yearly":
		return RecurYearly
	default:
		return RecurNone
	}
}

// NextDueDate calculates the next due date based on the task's recurrence
// frequency and interval. If the task has no DueDate, the current time is used
// as the base. The interval acts as a multiplier (e.g., RecurWeekly with
// interval 2 means every two weeks). An interval of 0 is treated as 1.
func NextDueDate(t Task) time.Time {
	base := time.Now()
	if t.DueDate != nil {
		base = *t.DueDate
	}

	interval := t.RecurInterval
	if interval <= 0 {
		interval = 1
	}

	switch t.RecurFreq {
	case RecurDaily:
		return base.AddDate(0, 0, interval)
	case RecurWeekly:
		return base.AddDate(0, 0, 7*interval)
	case RecurMonthly:
		return base.AddDate(0, interval, 0)
	case RecurYearly:
		return base.AddDate(interval, 0, 0)
	default:
		return base
	}
}
