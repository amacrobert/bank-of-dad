package interest

import (
	"log"
	"time"

	"bank-of-dad/internal/store"
)

// Scheduler processes due interest accruals in the background.
type Scheduler struct {
	interestStore *store.InterestStore
}

// NewScheduler creates a new interest Scheduler.
func NewScheduler(interestStore *store.InterestStore) *Scheduler {
	return &Scheduler{interestStore: interestStore}
}

// Start begins the background interest processing goroutine.
func (s *Scheduler) Start(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Process immediately on start (catch any missed while down)
		s.ProcessDue()

		for {
			select {
			case <-ticker.C:
				s.ProcessDue()
			case <-stop:
				return
			}
		}
	}()
}

// ProcessDue finds and applies interest to all eligible children.
func (s *Scheduler) ProcessDue() {
	dues, err := s.interestStore.ListDueForInterest()
	if err != nil {
		log.Printf("Error listing children due for interest: %v", err)
		return
	}

	for _, due := range dues {
		if err := s.interestStore.ApplyInterest(due.ChildID, due.ParentID, due.InterestRateBps); err != nil {
			log.Printf("Error applying interest for child %d: %v", due.ChildID, err)
			continue
		}
		log.Printf("Applied interest for child %d: %d bps on %d cents",
			due.ChildID, due.InterestRateBps, due.BalanceCents)
	}
}
