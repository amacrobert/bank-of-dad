package interest

import (
	"log"
	"time"

	"bank-of-dad/internal/allowance"
	"bank-of-dad/internal/store"
)

// Scheduler processes due interest accruals in the background.
type Scheduler struct {
	interestStore         *store.InterestStore
	interestScheduleStore *store.InterestScheduleStore
}

// NewScheduler creates a new interest Scheduler.
func NewScheduler(interestStore *store.InterestStore) *Scheduler {
	return &Scheduler{interestStore: interestStore}
}

// SetInterestScheduleStore sets the interest schedule store for schedule-based processing.
func (s *Scheduler) SetInterestScheduleStore(iss *store.InterestScheduleStore) {
	s.interestScheduleStore = iss
}

// Start begins the background interest processing goroutine.
func (s *Scheduler) Start(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Process immediately on start (catch any missed while down)
		s.processTick()

		for {
			select {
			case <-ticker.C:
				s.processTick()
			case <-stop:
				return
			}
		}
	}()
}

// processTick chooses between schedule-based and legacy processing.
func (s *Scheduler) processTick() {
	if s.interestScheduleStore != nil {
		s.ProcessDueSchedules()
	} else {
		s.ProcessDue()
	}
}

// ProcessDue finds and applies interest to all eligible children (legacy monthly-only path).
func (s *Scheduler) ProcessDue() {
	dues, err := s.interestStore.ListDueForInterest()
	if err != nil {
		log.Printf("Error listing children due for interest: %v", err)
		return
	}

	for _, due := range dues {
		if err := s.interestStore.ApplyInterest(due.ChildID, due.ParentID, due.InterestRateBps, 12); err != nil {
			log.Printf("Error applying interest for child %d: %v", due.ChildID, err)
			continue
		}
		log.Printf("Applied interest for child %d: %d bps on %d cents",
			due.ChildID, due.InterestRateBps, due.BalanceCents)
	}
}

// frequencyPeriodsPerYear maps schedule frequency to the number of periods per year for proration.
func frequencyPeriodsPerYear(freq store.Frequency) int {
	switch freq {
	case store.FrequencyWeekly:
		return 52
	case store.FrequencyBiweekly:
		return 26
	case store.FrequencyMonthly:
		return 12
	default:
		return 12
	}
}

// ProcessDueSchedules processes interest accruals based on interest_schedules table.
func (s *Scheduler) ProcessDueSchedules() {
	now := time.Now().UTC()
	schedules, err := s.interestScheduleStore.ListDue(now)
	if err != nil {
		log.Printf("Error listing due interest schedules: %v", err)
		return
	}

	for _, sched := range schedules {
		// Get the interest rate for this child
		rateBps, err := s.interestStore.GetInterestRate(sched.ChildID)
		if err != nil {
			log.Printf("Error getting interest rate for child %d: %v", sched.ChildID, err)
			continue
		}

		if rateBps <= 0 {
			log.Printf("Skipping interest for child %d: rate is %d bps", sched.ChildID, rateBps)
			// Still advance next_run_at so we don't keep retrying
			s.advanceNextRun(&sched)
			continue
		}

		periodsPerYear := frequencyPeriodsPerYear(sched.Frequency)

		if err := s.interestStore.ApplyInterest(sched.ChildID, sched.ParentID, rateBps, periodsPerYear); err != nil {
			log.Printf("Error applying interest for child %d: %v", sched.ChildID, err)
			// Still advance next_run_at to avoid retrying zero-interest cases
			s.advanceNextRun(&sched)
			continue
		}

		log.Printf("Applied scheduled interest for child %d: %d bps, %d periods/year",
			sched.ChildID, rateBps, periodsPerYear)

		s.advanceNextRun(&sched)
	}
}

// advanceNextRun calculates and updates the next_run_at for a schedule after execution.
func (s *Scheduler) advanceNextRun(sched *store.InterestSchedule) {
	tmpSched := &store.AllowanceSchedule{
		Frequency:  sched.Frequency,
		DayOfWeek:  sched.DayOfWeek,
		DayOfMonth: sched.DayOfMonth,
	}
	nextRun := allowance.CalculateNextRunAfterExecution(tmpSched, *sched.NextRunAt)
	if err := s.interestScheduleStore.UpdateNextRunAt(sched.ID, nextRun); err != nil {
		log.Printf("Error updating next_run_at for interest schedule %d: %v", sched.ID, err)
	}
}
