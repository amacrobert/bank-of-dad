package chore

import (
	"log"
	"time"

	"bank-of-dad/models"
	"bank-of-dad/repositories"
)

// ChoreScheduler generates recurring chore instances and expires past-due ones.
type ChoreScheduler struct {
	choreRepo         *repositories.ChoreRepo
	choreInstanceRepo *repositories.ChoreInstanceRepo
	childRepo         *repositories.ChildRepo
	familyRepo        *repositories.FamilyRepo
}

// NewScheduler creates a new ChoreScheduler.
func NewScheduler(choreRepo *repositories.ChoreRepo, choreInstanceRepo *repositories.ChoreInstanceRepo, childRepo *repositories.ChildRepo, familyRepo *repositories.FamilyRepo) *ChoreScheduler {
	return &ChoreScheduler{
		choreRepo:         choreRepo,
		choreInstanceRepo: choreInstanceRepo,
		childRepo:         childRepo,
		familyRepo:        familyRepo,
	}
}

// Start begins the background chore scheduling goroutine.
func (s *ChoreScheduler) Start(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Process immediately on start
		s.GenerateInstances()
		s.ExpireInstances()

		for {
			select {
			case <-ticker.C:
				s.GenerateInstances()
				s.ExpireInstances()
			case <-stop:
				return
			}
		}
	}()
}

// CalculatePeriodBounds returns the start and end of the current period for a recurring chore.
// The period is timezone-aware using the provided location.
func CalculatePeriodBounds(recurrence models.ChoreRecurrence, dayOfWeek *int, dayOfMonth *int, now time.Time, loc *time.Location) (periodStart, periodEnd time.Time) {
	localNow := now.In(loc)

	switch recurrence {
	case models.ChoreRecurrenceDaily:
		start := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1).Add(-time.Second)
		return start, end

	case models.ChoreRecurrenceWeekly:
		dow := 0
		if dayOfWeek != nil {
			dow = *dayOfWeek
		}
		currentDow := int(localNow.Weekday())
		daysBack := (currentDow - dow + 7) % 7
		start := time.Date(localNow.Year(), localNow.Month(), localNow.Day()-daysBack, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 7).Add(-time.Second)
		return start, end

	case models.ChoreRecurrenceMonthly:
		start := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return start, end
	}

	return now, now
}

// GenerateInstances creates chore instances for all active recurring chores in the current period.
func (s *ChoreScheduler) GenerateInstances() {
	chores, err := s.choreRepo.ListAllActiveRecurring()
	if err != nil {
		log.Printf("Chore scheduler: error listing recurring chores: %v", err)
		return
	}

	now := time.Now().UTC()
	generated := 0

	for _, cwa := range chores {
		tz, err := s.familyRepo.GetTimezone(cwa.FamilyID)
		if err != nil {
			log.Printf("Chore scheduler: error getting timezone for family %d: %v", cwa.FamilyID, err)
			continue
		}
		loc := loadTimezone(tz)

		periodStart, periodEnd := CalculatePeriodBounds(cwa.Recurrence, cwa.DayOfWeek, cwa.DayOfMonth, now, loc)

		for _, childID := range cwa.ChildIDs {
			// Skip disabled children
			child, err := s.childRepo.GetByID(childID)
			if err != nil || child == nil || child.IsDisabled {
				continue
			}

			exists, err := s.choreInstanceRepo.ExistsForPeriod(cwa.ID, childID, periodStart)
			if err != nil {
				log.Printf("Chore scheduler: error checking period for chore %d child %d: %v", cwa.ID, childID, err)
				continue
			}
			if exists {
				continue
			}

			instance := &models.ChoreInstance{
				ChoreID:     cwa.ID,
				ChildID:     childID,
				RewardCents: cwa.RewardCents,
				Status:      models.ChoreInstanceStatusAvailable,
				PeriodStart: &periodStart,
				PeriodEnd:   &periodEnd,
			}
			if _, err := s.choreInstanceRepo.CreateInstance(instance); err != nil {
				log.Printf("Chore scheduler: error creating instance for chore %d child %d: %v", cwa.ID, childID, err)
				continue
			}
			generated++
		}
	}

	if generated > 0 {
		log.Printf("Chore scheduler: generated %d instances", generated)
	}
}

// ExpireInstances marks all available instances past their period_end as expired.
func (s *ChoreScheduler) ExpireInstances() {
	count, err := s.choreInstanceRepo.ExpireByPeriod(time.Now().UTC())
	if err != nil {
		log.Printf("Chore scheduler: error expiring instances: %v", err)
		return
	}
	if count > 0 {
		log.Printf("Chore scheduler: expired %d instances", count)
	}
}

func loadTimezone(tz string) *time.Location {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	return time.UTC
}
