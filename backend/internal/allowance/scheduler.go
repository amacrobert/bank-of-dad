package allowance

import (
	"log"
	"time"

	"bank-of-dad/internal/store"
)

// Scheduler processes due allowance schedules in the background.
type Scheduler struct {
	scheduleStore *store.ScheduleStore
	txStore       *store.TransactionStore
	childStore    *store.ChildStore
}

// NewScheduler creates a new Scheduler.
func NewScheduler(scheduleStore *store.ScheduleStore, txStore *store.TransactionStore, childStore *store.ChildStore) *Scheduler {
	return &Scheduler{
		scheduleStore: scheduleStore,
		txStore:       txStore,
		childStore:    childStore,
	}
}

// Start begins the background schedule processing goroutine.
func (s *Scheduler) Start(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Process immediately on start (catch any missed while down)
		s.ProcessDueSchedules()

		for {
			select {
			case <-ticker.C:
				s.ProcessDueSchedules()
			case <-stop:
				return
			}
		}
	}()
}

// ProcessDueSchedules finds and executes all schedules that are due.
func (s *Scheduler) ProcessDueSchedules() {
	schedules, err := s.scheduleStore.ListDue(time.Now())
	if err != nil {
		log.Printf("Error listing due schedules: %v", err)
		return
	}

	for _, sched := range schedules {
		if err := s.executeSchedule(sched); err != nil {
			log.Printf("Error executing schedule %d: %v", sched.ID, err)
		}
	}
}

// executeSchedule creates a deposit transaction and advances the schedule's next_run_at.
func (s *Scheduler) executeSchedule(sched store.AllowanceSchedule) error {
	// Build note from schedule
	var note string
	if sched.Note != nil {
		note = *sched.Note
	}

	// Create allowance deposit
	_, _, err := s.txStore.DepositAllowance(
		sched.ChildID,
		sched.ParentID,
		sched.AmountCents,
		sched.ID,
		note,
	)
	if err != nil {
		return err
	}

	// Calculate and set next run time
	executedAt := time.Now().UTC()
	if sched.NextRunAt != nil {
		executedAt = *sched.NextRunAt
	}
	nextRun := CalculateNextRunAfterExecution(&sched, executedAt)

	if err := s.scheduleStore.UpdateNextRunAt(sched.ID, nextRun); err != nil {
		return err
	}

	log.Printf("Executed schedule %d: deposited %d cents to child %d, next run: %s",
		sched.ID, sched.AmountCents, sched.ChildID, nextRun.Format(time.DateOnly))

	return nil
}
