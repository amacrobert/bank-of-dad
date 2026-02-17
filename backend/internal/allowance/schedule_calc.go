package allowance

import (
	"time"

	"bank-of-dad/internal/store"
)

// CalculateNextRun determines the next run time for a schedule based on its frequency.
// Used when creating a new schedule or resuming a paused one.
// The loc parameter specifies the family's timezone — next_run_at is computed as
// midnight in that timezone, returned as the equivalent UTC time.
func CalculateNextRun(sched *store.AllowanceSchedule, after time.Time, loc *time.Location) time.Time {
	switch sched.Frequency {
	case store.FrequencyWeekly:
		return nextWeeklyDate(*sched.DayOfWeek, after, loc)
	case store.FrequencyBiweekly:
		return nextBiweeklyDate(*sched.DayOfWeek, after, loc)
	case store.FrequencyMonthly:
		return nextMonthlyDate(*sched.DayOfMonth, after, loc)
	}
	return after
}

// CalculateNextRunAfterExecution determines the next run time after a schedule has just executed.
// For weekly: next week same day. For biweekly: 14 days later. For monthly: same day next month.
// The loc parameter specifies the family's timezone for midnight calculation.
func CalculateNextRunAfterExecution(sched *store.AllowanceSchedule, executedAt time.Time, loc *time.Location) time.Time {
	// Convert executedAt to the family's timezone to get the correct calendar date
	localExecuted := executedAt.In(loc)
	switch sched.Frequency {
	case store.FrequencyWeekly:
		next := localExecuted.AddDate(0, 0, 7)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, loc)
	case store.FrequencyBiweekly:
		next := localExecuted.AddDate(0, 0, 14)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, loc)
	case store.FrequencyMonthly:
		return nextMonthlyDate(*sched.DayOfMonth, executedAt, loc)
	}
	return executedAt
}

// nextWeeklyDate returns the next occurrence of the given weekday strictly after 'after'.
// The result is midnight in the given timezone on that day.
func nextWeeklyDate(dayOfWeek int, after time.Time, loc *time.Location) time.Time {
	after = after.In(loc)
	daysUntil := (dayOfWeek - int(after.Weekday()) + 7) % 7
	if daysUntil == 0 {
		daysUntil = 7 // If today is the day, go to next week
	}
	return time.Date(after.Year(), after.Month(), after.Day()+daysUntil,
		0, 0, 0, 0, loc)
}

// nextBiweeklyDate returns the next matching weekday. If after is already on that day,
// it advances 14 days. Otherwise it finds the next occurrence (which will be the first one).
// The result is midnight in the given timezone on that day.
func nextBiweeklyDate(dayOfWeek int, after time.Time, loc *time.Location) time.Time {
	after = after.In(loc)
	daysUntil := (dayOfWeek - int(after.Weekday()) + 7) % 7
	if daysUntil == 0 {
		daysUntil = 14 // Same day → 14 days out
	}
	return time.Date(after.Year(), after.Month(), after.Day()+daysUntil,
		0, 0, 0, 0, loc)
}

// nextMonthlyDate returns the next occurrence of the given day-of-month strictly after 'after'.
// If the target day exceeds the days in a month, it clamps to the last day.
// The result is midnight in the given timezone on that day.
func nextMonthlyDate(dayOfMonth int, after time.Time, loc *time.Location) time.Time {
	after = after.In(loc)
	year, month, day := after.Date()

	// Try this month first
	clampedDay := min(dayOfMonth, daysInMonth(year, month))
	target := time.Date(year, month, clampedDay, 0, 0, 0, 0, loc)
	if target.After(after) || (clampedDay > day) {
		return target
	}

	// Otherwise next month
	month++
	if month > 12 {
		month = 1
		year++
	}
	clampedDay = min(dayOfMonth, daysInMonth(year, month))
	return time.Date(year, month, clampedDay, 0, 0, 0, 0, loc)
}

// daysInMonth returns the number of days in the given month.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
