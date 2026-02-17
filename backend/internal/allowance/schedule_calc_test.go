package allowance

import (
	"testing"
	"time"

	"bank-of-dad/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func dateIn(year int, month time.Month, day int, loc *time.Location) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func intPtr(i int) *int {
	return &i
}

// === nextWeeklyDate tests (UTC) ===

func TestNextWeeklyDate_NormalCase(t *testing.T) {
	// Wednesday 2026-02-04, looking for Friday (5)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(5, after, time.UTC) // Friday
	assert.Equal(t, date(2026, time.February, 6), result)
}

func TestNextWeeklyDate_SameDayGoesToNextWeek(t *testing.T) {
	// Friday 2026-02-06, looking for Friday (5)
	after := date(2026, time.February, 6)
	result := nextWeeklyDate(5, after, time.UTC) // Friday
	assert.Equal(t, date(2026, time.February, 13), result)
}

func TestNextWeeklyDate_Sunday(t *testing.T) {
	// Wednesday 2026-02-04, looking for Sunday (0)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(0, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 8), result)
}

func TestNextWeeklyDate_Monday(t *testing.T) {
	// Wednesday 2026-02-04, looking for Monday (1)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(1, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 9), result)
}

func TestNextWeeklyDate_Tomorrow(t *testing.T) {
	// Wednesday 2026-02-04, looking for Thursday (4)
	after := date(2026, time.February, 4)
	result := nextWeeklyDate(4, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 5), result)
}

// === nextBiweeklyDate tests (UTC) ===

func TestNextBiweeklyDate_14DaysFromAfter(t *testing.T) {
	// Friday 2026-02-06, looking for next biweekly Friday
	after := date(2026, time.February, 6)
	result := nextBiweeklyDate(5, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 20), result)
}

func TestNextBiweeklyDate_SameDay(t *testing.T) {
	// If after is a Friday and day_of_week is Friday, next occurrence is in 14 days
	after := date(2026, time.February, 6) // Friday
	result := nextBiweeklyDate(5, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 20), result)
}

func TestNextBiweeklyDate_DifferentDay(t *testing.T) {
	// Wednesday 2026-02-04, looking for biweekly Friday
	after := date(2026, time.February, 4)
	result := nextBiweeklyDate(5, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 6), result)
}

// === nextMonthlyDate tests (UTC) ===

func TestNextMonthlyDate_FutureThisMonth(t *testing.T) {
	// Feb 4, target day 15
	after := date(2026, time.February, 4)
	result := nextMonthlyDate(15, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 15), result)
}

func TestNextMonthlyDate_PastThisMonth(t *testing.T) {
	// Feb 15, target day 1 → goes to March 1
	after := date(2026, time.February, 15)
	result := nextMonthlyDate(1, after, time.UTC)
	assert.Equal(t, date(2026, time.March, 1), result)
}

func TestNextMonthlyDate_SameDay(t *testing.T) {
	// Feb 15, target day 15 → goes to March 15
	after := date(2026, time.February, 15)
	result := nextMonthlyDate(15, after, time.UTC)
	assert.Equal(t, date(2026, time.March, 15), result)
}

func TestNextMonthlyDate_EndOfMonthClamping_31InFeb(t *testing.T) {
	// Jan 15, target day 31 → Jan 31
	after := date(2026, time.January, 15)
	result := nextMonthlyDate(31, after, time.UTC)
	assert.Equal(t, date(2026, time.January, 31), result)

	// Feb 1, target day 31 → Feb 28 (2026 is not a leap year)
	after = date(2026, time.February, 1)
	result = nextMonthlyDate(31, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 28), result)
}

func TestNextMonthlyDate_EndOfMonthClamping_31InApr(t *testing.T) {
	// March 31, target day 31 → April 30
	after := date(2026, time.March, 31)
	result := nextMonthlyDate(31, after, time.UTC)
	assert.Equal(t, date(2026, time.April, 30), result)
}

func TestNextMonthlyDate_EndOfMonthClamping_31InJun(t *testing.T) {
	// May 31, target day 31 → June 30
	after := date(2026, time.May, 31)
	result := nextMonthlyDate(31, after, time.UTC)
	assert.Equal(t, date(2026, time.June, 30), result)
}

func TestNextMonthlyDate_LeapYear(t *testing.T) {
	// Feb 29 in leap year 2028
	after := date(2028, time.February, 1)
	result := nextMonthlyDate(29, after, time.UTC)
	assert.Equal(t, date(2028, time.February, 29), result)

	// Feb 29 in non-leap year 2026 → Feb 28
	after = date(2026, time.February, 1)
	result = nextMonthlyDate(29, after, time.UTC)
	assert.Equal(t, date(2026, time.February, 28), result)
}

func TestNextMonthlyDate_DecemberToJanuary(t *testing.T) {
	// Dec 25, target day 1 → Jan 1 next year
	after := date(2026, time.December, 25)
	result := nextMonthlyDate(1, after, time.UTC)
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

// === calculateNextRun tests (UTC) ===

func TestCalculateNextRun_Weekly(t *testing.T) {
	now := date(2026, time.February, 4) // Wednesday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(5), // Friday
	}
	result := CalculateNextRun(sched, now, time.UTC)
	assert.Equal(t, date(2026, time.February, 6), result)
}

func TestCalculateNextRun_Biweekly(t *testing.T) {
	now := date(2026, time.February, 6) // Friday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyBiweekly,
		DayOfWeek: intPtr(5), // Friday
	}
	result := CalculateNextRun(sched, now, time.UTC)
	assert.Equal(t, date(2026, time.February, 20), result)
}

func TestCalculateNextRun_Monthly(t *testing.T) {
	now := date(2026, time.February, 4)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(15),
	}
	result := CalculateNextRun(sched, now, time.UTC)
	assert.Equal(t, date(2026, time.February, 15), result)
}

// === CalculateNextRunAfterExecution tests (UTC) ===

func TestCalculateNextRunAfterExecution_Weekly(t *testing.T) {
	executedAt := date(2026, time.February, 6) // Friday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(5),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt, time.UTC)
	assert.Equal(t, date(2026, time.February, 13), result) // Next Friday
}

func TestCalculateNextRunAfterExecution_Biweekly(t *testing.T) {
	executedAt := date(2026, time.February, 6) // Friday
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyBiweekly,
		DayOfWeek: intPtr(5),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt, time.UTC)
	assert.Equal(t, date(2026, time.February, 20), result) // 14 days later
}

func TestCalculateNextRunAfterExecution_Monthly(t *testing.T) {
	executedAt := date(2026, time.January, 15)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(15),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt, time.UTC)
	assert.Equal(t, date(2026, time.February, 15), result)
}

func TestCalculateNextRunAfterExecution_Monthly_31st(t *testing.T) {
	executedAt := date(2026, time.January, 31)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(31),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt, time.UTC)
	// Feb 2026 has 28 days
	assert.Equal(t, date(2026, time.February, 28), result)
}

// =============================================================================
// Timezone-aware tests
// =============================================================================

func TestNextWeeklyDate_NewYork(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Tuesday Feb 17, 2026 in New York, looking for Wednesday (3)
	after := dateIn(2026, time.February, 17, loc)
	result := nextWeeklyDate(3, after, loc)

	// Should be Wednesday Feb 18 at midnight New York time
	expected := dateIn(2026, time.February, 18, loc)
	assert.Equal(t, expected, result)

	// Verify the UTC equivalent: midnight EST = 5am UTC
	assert.Equal(t, 5, result.UTC().Hour())
	assert.Equal(t, 18, result.UTC().Day())
}

func TestNextWeeklyDate_LosAngeles(t *testing.T) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	// Tuesday Feb 17, 2026 in LA, looking for Wednesday (3)
	after := dateIn(2026, time.February, 17, loc)
	result := nextWeeklyDate(3, after, loc)

	// Should be Wednesday Feb 18 at midnight Pacific time
	expected := dateIn(2026, time.February, 18, loc)
	assert.Equal(t, expected, result)

	// Verify the UTC equivalent: midnight PST = 8am UTC
	assert.Equal(t, 8, result.UTC().Hour())
	assert.Equal(t, 18, result.UTC().Day())
}

func TestNextWeeklyDate_Kolkata(t *testing.T) {
	// Non-hour offset timezone: UTC+5:30
	loc, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	// Tuesday Feb 17 in Kolkata, looking for Wednesday (3)
	after := dateIn(2026, time.February, 17, loc)
	result := nextWeeklyDate(3, after, loc)

	expected := dateIn(2026, time.February, 18, loc)
	assert.Equal(t, expected, result)

	// Verify UTC: midnight IST = 6:30pm previous day UTC
	assert.Equal(t, 17, result.UTC().Day())
	assert.Equal(t, 18, result.UTC().Hour())
	assert.Equal(t, 30, result.UTC().Minute())
}

func TestNextMonthlyDate_NewYork(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Feb 4 in New York, target day 1 → March 1
	after := dateIn(2026, time.February, 4, loc)
	result := nextMonthlyDate(1, after, loc)

	expected := dateIn(2026, time.March, 1, loc)
	assert.Equal(t, expected, result)

	// Verify UTC: midnight EST = 5am UTC
	assert.Equal(t, 5, result.UTC().Hour())
	assert.Equal(t, 1, result.UTC().Day())
}

func TestCalculateNextRun_Weekly_NewYork(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Tuesday Feb 17, 2026 at 3pm New York, schedule for Wednesday (3)
	now := time.Date(2026, time.February, 17, 15, 0, 0, 0, loc)
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(3), // Wednesday
	}
	result := CalculateNextRun(sched, now, loc)

	// Should be Wednesday Feb 18 at midnight New York
	expected := dateIn(2026, time.February, 18, loc)
	assert.Equal(t, expected, result)
	assert.Equal(t, 5, result.UTC().Hour()) // midnight EST = 5am UTC
}

func TestCalculateNextRunAfterExecution_Weekly_NewYork(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Executed at midnight EST Wednesday Feb 18
	executedAt := dateIn(2026, time.February, 18, loc)
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(3), // Wednesday
	}
	result := CalculateNextRunAfterExecution(sched, executedAt, loc)

	// Should be next Wednesday Feb 25 at midnight EST
	expected := dateIn(2026, time.February, 25, loc)
	assert.Equal(t, expected, result)
	assert.Equal(t, 5, result.UTC().Hour())
}

func TestCalculateNextRunAfterExecution_Monthly_NewYork(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Executed at midnight EST on the 15th
	executedAt := dateIn(2026, time.January, 15, loc)
	sched := &store.AllowanceSchedule{
		Frequency:  store.FrequencyMonthly,
		DayOfMonth: intPtr(15),
	}
	result := CalculateNextRunAfterExecution(sched, executedAt, loc)

	// Should be Feb 15 at midnight EST
	expected := dateIn(2026, time.February, 15, loc)
	assert.Equal(t, expected, result)
	assert.Equal(t, 5, result.UTC().Hour())
}

// === DST Transition Tests ===

func TestNextWeeklyDate_DSTSpringForward(t *testing.T) {
	// 2026 spring forward: March 8 at 2am EST → 3am EDT
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// March 7, 2026 (Saturday) in New York, looking for Sunday (0) = March 8
	after := dateIn(2026, time.March, 7, loc)
	result := nextWeeklyDate(0, after, loc)

	// Should be Sunday March 8 at midnight EDT (which is 4am UTC because clocks spring forward)
	// Actually March 8 midnight is still EST (2am hasn't happened yet), so 5am UTC
	expected := dateIn(2026, time.March, 8, loc)
	assert.Equal(t, expected, result)
	// Midnight on March 8 is still EST (before 2am), so 5am UTC
	assert.Equal(t, 5, result.UTC().Hour())

	// Now test a date AFTER DST change: March 14 (Saturday), looking for Sunday (0) = March 15
	after = dateIn(2026, time.March, 14, loc)
	result = nextWeeklyDate(0, after, loc)

	// March 15 is now EDT (UTC-4), so midnight EDT = 4am UTC
	expected = dateIn(2026, time.March, 15, loc)
	assert.Equal(t, expected, result)
	assert.Equal(t, 4, result.UTC().Hour()) // Now EDT
}

func TestNextWeeklyDate_DSTFallBack(t *testing.T) {
	// 2026 fall back: November 1 at 2am EDT → 1am EST
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Oct 31, 2026 (Saturday) in New York, looking for Sunday (0) = Nov 1
	after := dateIn(2026, time.October, 31, loc)
	result := nextWeeklyDate(0, after, loc)

	// November 1, midnight EDT (before fall back at 2am) = 4am UTC
	expected := dateIn(2026, time.November, 1, loc)
	assert.Equal(t, expected, result)

	// Nov 7, 2026 (Saturday), looking for Sunday (0) = Nov 8
	after = dateIn(2026, time.November, 7, loc)
	result = nextWeeklyDate(0, after, loc)

	// November 8 is now EST (UTC-5), so midnight EST = 5am UTC
	expected = dateIn(2026, time.November, 8, loc)
	assert.Equal(t, expected, result)
	assert.Equal(t, 5, result.UTC().Hour())
}

func TestCalculateNextRun_UTCInput_NonUTCLocation(t *testing.T) {
	// Simulate the handler scenario: time.Now().UTC() is passed as 'after'
	// but location is America/New_York
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// 3pm UTC on Tuesday Feb 17 = 10am EST on Tuesday Feb 17
	now := time.Date(2026, time.February, 17, 15, 0, 0, 0, time.UTC)
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(3), // Wednesday
	}
	result := CalculateNextRun(sched, now, loc)

	// In EST, it's Tuesday 10am → next Wednesday is Feb 18
	expected := dateIn(2026, time.February, 18, loc)
	assert.Equal(t, expected, result)
}

func TestCalculateNextRun_UTCTimezoneMatchesOldBehavior(t *testing.T) {
	// When loc = time.UTC, behavior should match the old UTC-only code
	now := date(2026, time.February, 4) // Wednesday midnight UTC
	sched := &store.AllowanceSchedule{
		Frequency: store.FrequencyWeekly,
		DayOfWeek: intPtr(5), // Friday
	}
	result := CalculateNextRun(sched, now, time.UTC)
	assert.Equal(t, date(2026, time.February, 6), result)
	assert.Equal(t, 0, result.UTC().Hour()) // Midnight UTC
}
