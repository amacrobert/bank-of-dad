package allowance

import (
	"testing"
	"time"

	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/assert"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func intPtr(i int) *int {
	return &i
}

// === nextWeeklyDate tests ===

func TestNextWeeklyDate_NormalCase(t *testing.T) {
	// Wednesday 2026-02-04, looking for Friday (5)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(5, after) // Friday
	assert.Equal(t, date(2026, time.February, 6), result)
}

func TestNextWeeklyDate_SameDayGoesToNextWeek(t *testing.T) {
	// Friday 2026-02-06, looking for Friday (5)
	after := date(2026, time.February, 6)
	result := nextWeeklyDate(5, after) // Friday
	assert.Equal(t, date(2026, time.February, 13), result)
}

func TestNextWeeklyDate_Sunday(t *testing.T) {
	// Wednesday 2026-02-04, looking for Sunday (0)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(0, after)
	assert.Equal(t, date(2026, time.February, 8), result)
}

func TestNextWeeklyDate_Monday(t *testing.T) {
	// Wednesday 2026-02-04, looking for Monday (1)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(1, after)
	assert.Equal(t, date(2026, time.February, 9), result)
}

func TestNextWeeklyDate_Tomorrow(t *testing.T) {
	// Wednesday 2026-02-04, looking for Thursday (4)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(4, after)
	assert.Equal(t, date(2026, time.February, 5), result)
}

// === nextBiweeklyDate tests ===

func TestNextBiweeklyDate_14DaysFromAfter(t *testing.T) {
	// Friday 2026-02-06, looking for next biweekly Friday
	after := date(2026, time.February, 6)
	result := nextBiweeklyDate(5, after)
	assert.Equal(t, date(2026, time.February, 20), result)
}

func TestNextBiweeklyDate_SameDay(t *testing.T) {
	// If after is a Friday and day_of_week is Friday, next occurrence is in 14 days
	after := date(2026, time.February, 6) // Friday
	result := nextBiweeklyDate(5, after)
	assert.Equal(t, date(2026, time.February, 20), result)
}

func TestNextBiweeklyDate_DifferentDay(t *testing.T) {
	// Wednesday 2026-02-04, looking for biweekly Friday
	after := date(2026, time.February, 4)
	result := nextBiweeklyDate(5, after)
	// Should go to next Friday (Feb 6) + 7 days = Feb 13? No, biweekly means next matching day + 14 days from that
	// Actually biweekly: find next matching day, that's the first one
	assert.Equal(t, date(2026, time.February, 6), result)
}

// === nextMonthlyDate tests ===

func TestNextMonthlyDate_FutureThisMonth(t *testing.T) {
	// Feb 4, target day 15
	after := date(2026, time.February, 4)
	result := nextMonthlyDate(15, after)
	assert.Equal(t, date(2026, time.February, 15), result)
}

func TestNextMonthlyDate_PastThisMonth(t *testing.T) {
	// Feb 15, target day 1 → goes to March 1
	after := date(2026, time.February, 15)
	result := nextMonthlyDate(1, after)
	assert.Equal(t, date(2026, time.March, 1), result)
}

func TestNextMonthlyDate_SameDay(t *testing.T) {
	// Feb 15, target day 15 → goes to March 15
	after := date(2026, time.February, 15)
	result := nextMonthlyDate(15, after)
	assert.Equal(t, date(2026, time.March, 15), result)
}

func TestNextMonthlyDate_EndOfMonthClamping_31InFeb(t *testing.T) {
	// Jan 15, target day 31 → Jan 31
	after := date(2026, time.January, 15)
	result := nextMonthlyDate(31, after)
	assert.Equal(t, date(2026, time.January, 31), result)

	// Feb 1, target day 31 → Feb 28 (2026 is not a leap year)
	after = date(2026, time.February, 1)
	result = nextMonthlyDate(31, after)
	assert.Equal(t, date(2026, time.February, 28), result)
}

func TestNextMonthlyDate_EndOfMonthClamping_31InApr(t *testing.T) {
	// March 31, target day 31 → April 30
	after := date(2026, time.March, 31)
	result := nextMonthlyDate(31, after)
	assert.Equal(t, date(2026, time.April, 30), result)
}

func TestNextMonthlyDate_EndOfMonthClamping_31InJun(t *testing.T) {
	// May 31, target day 31 → June 30
	after := date(2026, time.May, 31)
	result := nextMonthlyDate(31, after)
	assert.Equal(t, date(2026, time.June, 30), result)
}

func TestNextMonthlyDate_LeapYear(t *testing.T) {
	// Feb 29 in leap year 2028
	after := date(2028, time.February, 1)
	result := nextMonthlyDate(29, after)
	assert.Equal(t, date(2028, time.February, 29), result)

	// Feb 29 in non-leap year 2026 → Feb 28
	after = date(2026, time.February, 1)
	result = nextMonthlyDate(29, after)
	assert.Equal(t, date(2026, time.February, 28), result)
}

func TestNextMonthlyDate_DecemberToJanuary(t *testing.T) {
	// Dec 25, target day 1 → Jan 1 next year
	after := date(2026, time.December, 25)
	result := nextMonthlyDate(1, after)
	assert.Equal(t, date(2027, time.January, 1), result)
}

// === daysInMonth tests ===

func TestDaysInMonth(t *testing.T) {
	assert.Equal(t, 31, daysInMonth(2026, time.January))
	assert.Equal(t, 28, daysInMonth(2026, time.February))
	assert.Equal(t, 29, daysInMonth(2028, time.February)) // leap year
	assert.Equal(t, 31, daysInMonth(2026, time.March))
	assert.Equal(t, 30, daysInMonth(2026, time.April))
	assert.Equal(t, 30, daysInMonth(2026, time.June))
	assert.Equal(t, 30, daysInMonth(2026, time.September))
	assert.Equal(t, 30, daysInMonth(2026, time.November))
	assert.Equal(t, 31, daysInMonth(2026, time.December))
}

// === calculateNextRun tests ===

func TestCalculateNextRun_Weekly(t *testing.T) {
	now := date(2026, time.February, 4) // Wednesday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(5), // Friday
	}
	result := CalculateNextRun(sched, now)
	assert.Equal(t, date(2026, time.February, 6), result)
}

func TestCalculateNextRun_Biweekly(t *testing.T) {
	now := date(2026, time.February, 6) // Friday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyBiweekly,
		DayOfWeek: intPtr(5), // Friday
	}
	result := CalculateNextRun(sched, now)
	assert.Equal(t, date(2026, time.February, 20), result)
}

func TestCalculateNextRun_Monthly(t *testing.T) {
	now := date(2026, time.February, 4)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(15),
	}
	result := CalculateNextRun(sched, now)
	assert.Equal(t, date(2026, time.February, 15), result)
}

// === CalculateNextRunAfterExecution tests ===

func TestCalculateNextRunAfterExecution_Weekly(t *testing.T) {
	executedAt := date(2026, time.February, 6) // Friday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(5),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt)
	assert.Equal(t, date(2026, time.February, 13), result) // Next Friday
}

func TestCalculateNextRunAfterExecution_Biweekly(t *testing.T) {
	executedAt := date(2026, time.February, 6) // Friday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyBiweekly,
		DayOfWeek: intPtr(5),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt)
	assert.Equal(t, date(2026, time.February, 20), result) // 14 days later
}

func TestCalculateNextRunAfterExecution_Monthly(t *testing.T) {
	executedAt := date(2026, time.January, 15)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(15),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt)
	assert.Equal(t, date(2026, time.February, 15), result)
}

func TestCalculateNextRunAfterExecution_Monthly_31st(t *testing.T) {
	executedAt := date(2026, time.January, 31)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(31),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt)
	// Feb 2026 has 28 days
	assert.Equal(t, date(2026, time.February, 28), result)
}
