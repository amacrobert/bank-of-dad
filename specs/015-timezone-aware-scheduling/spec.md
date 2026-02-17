# Feature Specification: Timezone-Aware Scheduling

**Feature Branch**: `015-timezone-aware-scheduling`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "As a user, I want all dates to be relative to my family's timezone and I want dates to appear as expected. Interest payments and allowance payments should happen as early as possible on the day they're due, taking into account the family's timezone when determining that date."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scheduled Payments Happen at the Right Time in My Timezone (Priority: P1)

As a parent, I want allowance and interest payments to be distributed at midnight (or shortly after) in my family's configured timezone on the scheduled day, so that my children see their money arrive on the correct calendar date.

Currently, payments are triggered based on UTC midnight, which means a family in the US Eastern timezone (UTC-5) sees payments arrive at 7pm the evening before the scheduled day. This causes confusion because the "upcoming" date and the actual distribution don't align with the family's local calendar.

**Why this priority**: This is the core bug driving this feature. Payments arriving on the wrong calendar day undermines trust in the system and causes confusion for both parents and children.

**Independent Test**: Can be tested by configuring a family's timezone, setting an allowance schedule to a specific day, and verifying the payment is distributed at midnight in that timezone rather than midnight UTC.

**Acceptance Scenarios**:

1. **Given** a family with timezone set to "America/New_York" and a weekly allowance scheduled for Wednesday, **When** it is 12:00am Wednesday in New York (5:00am UTC Wednesday), **Then** the allowance payment is distributed at that time (not at 12:00am UTC, which is 7:00pm Tuesday in New York).
2. **Given** a family with timezone set to "America/Los_Angeles" and a monthly interest payment scheduled for the 1st, **When** it is 12:00am on the 1st in Los Angeles (8:00am UTC on the 1st), **Then** the interest payment is distributed at that time.
3. **Given** a family with timezone set to "America/New_York" during daylight saving time (UTC-4), **When** a scheduled payment is due, **Then** the system correctly accounts for the DST offset and distributes at 12:00am EDT (4:00am UTC).
4. **Given** a family whose timezone is changed from "America/New_York" to "America/Chicago", **When** the next scheduled payment is due, **Then** the payment is distributed at midnight in the new timezone (Central Time).

---

### User Story 2 - Upcoming Dates Display Correctly in My Timezone (Priority: P1)

As a parent, when I view upcoming allowance or interest payment dates, I want those dates to reflect my family's timezone so the displayed day matches what I expect on my local calendar.

Currently, the "Upcoming" date can display a day early because the next-due calculation uses UTC. For example, on Tuesday in New York, a Wednesday allowance shows "Feb 17" (Tuesday's UTC date) instead of "Feb 18" (Wednesday).

**Why this priority**: This is directly tied to the same root cause as Story 1 and is equally critical — showing the wrong date erodes user trust even if the payment technically fires on the right UTC timestamp.

**Independent Test**: Can be tested by setting a family timezone, creating a schedule for a specific day, and verifying the displayed "upcoming" date matches the expected local calendar date.

**Acceptance Scenarios**:

1. **Given** it is Tuesday Feb 17, 2026 in the America/New_York timezone and a weekly allowance is set for Wednesday, **When** the parent views the upcoming allowance date, **Then** the displayed date is "Feb 18, 2026" (Wednesday), not "Feb 17, 2026" (Tuesday).
2. **Given** a family in the America/Los_Angeles timezone and a monthly interest payment due on the 1st, **When** the parent views upcoming interest date on the last day of the month, **Then** the displayed date shows the 1st of the next month (in local time).
3. **Given** a schedule with no upcoming payment (e.g., paused or inactive), **When** the parent views the schedule, **Then** no misleading date is shown.

---

### User Story 3 - All Date Displays Are Timezone-Consistent (Priority: P2)

As a parent or child, when I view transaction dates, schedule summaries, or any date shown in the application, I want all dates to be displayed relative to my family's timezone so everything is consistent.

**Why this priority**: While Stories 1 and 2 address the core scheduling and "upcoming" display bugs, other date displays (transaction history, last-paid dates, etc.) should also be consistent. This prevents confusion where a payment distributed at midnight Eastern shows a previous day's date in transaction history.

**Independent Test**: Can be tested by creating transactions in various timezones and verifying all displayed dates match the family's configured timezone.

**Acceptance Scenarios**:

1. **Given** a family in America/New_York and an allowance distributed at 12:00am Wednesday Eastern (5:00am Wednesday UTC), **When** the parent or child views the transaction in their history, **Then** the transaction date shows Wednesday's date, not Tuesday's.
2. **Given** a family in Pacific timezone (UTC-8) and a transaction that occurred at 11:00pm Pacific on March 15, **When** the user views the transaction, **Then** the date displays as March 15, not March 16 (even though it's already March 16 in UTC).

---

### Edge Cases

- What happens when a family has not configured a timezone? The system falls back to UTC and functions correctly.
- What happens during daylight saving time transitions (spring forward / fall back)? The system handles the shift so payments aren't skipped or doubled. On "spring forward" night when 2:00am becomes 3:00am, midnight still exists and payments fire normally. On "fall back" night, the system does not distribute a payment twice.
- What happens if a payment was due while the system was down? Existing catch-up behavior continues to work — overdue payments are distributed when the system comes back online.
- What happens for timezones with non-hour offsets (e.g., India at UTC+5:30, Nepal at UTC+5:45)? The system supports all standard IANA timezones.
- What happens when a family changes their timezone between when a payment is scheduled and when it's due? The payment uses the family's timezone at the time of evaluation (current setting), not the timezone when the schedule was created.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST use the family's configured timezone when determining whether a scheduled allowance or interest payment is due.
- **FR-002**: The system MUST distribute scheduled payments at or shortly after midnight in the family's configured timezone on the scheduled day.
- **FR-003**: The system MUST display "upcoming" payment dates that correspond to the correct calendar date in the family's timezone.
- **FR-004**: The system MUST display all transaction dates relative to the family's configured timezone.
- **FR-005**: The system MUST correctly handle daylight saving time transitions without skipping or duplicating payments.
- **FR-006**: The system MUST support all standard IANA timezone identifiers (e.g., "America/New_York", "Europe/London", "Asia/Tokyo").
- **FR-007**: The system MUST fall back to UTC if a family has no timezone configured.
- **FR-008**: The system MUST use the family's current timezone setting when evaluating payment due dates (not the timezone at the time the schedule was originally created).
- **FR-009**: The system MUST NOT change the behavior of manual (ad-hoc) transactions — only scheduled/automated payments and date displays are affected.

### Key Entities

- **Family**: Has a configured timezone (IANA timezone identifier, e.g., "America/New_York"). This is the authoritative source for all timezone-related decisions for all members of the family.
- **Allowance Schedule**: Defines recurring allowance payments (weekly on a day, or monthly on a date). The "next due" evaluation and "upcoming" display must use the family's timezone.
- **Interest Schedule**: Defines recurring interest calculations (weekly or monthly). Same timezone-awareness requirements as allowance schedules.
- **Transaction**: Records of payments made. The displayed date should reflect the family's timezone.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of scheduled payments are distributed on the correct calendar date as perceived by the family in their configured timezone.
- **SC-002**: 100% of "upcoming" payment dates displayed in the application match the expected calendar date in the family's timezone.
- **SC-003**: All transaction dates displayed in the application reflect the family's configured timezone, with zero discrepancy between the displayed date and the family's local calendar date.
- **SC-004**: Daylight saving time transitions cause zero skipped or duplicated payments across all configured timezones.
- **SC-005**: Families can use any standard IANA timezone and receive correct scheduling behavior.

## Assumptions

- The family's timezone is already stored in the database (the `families` table has a `timezone` column, added in feature 013-parent-settings).
- The existing scheduler runs on a periodic interval and checks for due payments. This feature modifies *how* "due" is determined, not the scheduler's fundamental architecture.
- The system currently uses UTC for all date/time calculations. This feature shifts the "due date" evaluation to be timezone-relative while keeping internal storage in UTC.
- Manual/ad-hoc transactions entered by parents are not affected by this change — they are recorded at the time they are submitted.
- The frontend currently displays dates without timezone conversion. This feature requires the frontend to render dates in the family's timezone.
