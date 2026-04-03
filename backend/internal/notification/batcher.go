package notification

import (
	"log"
	"sync"
	"time"
)

const batchWindow = 5 * time.Minute

// batchKey uniquely identifies a batch by family and child.
type batchKey struct {
	FamilyID int64
	ChildID  int64
}

// batchEntry holds accumulated chore completions for a family+child pair.
type batchEntry struct {
	ChildName string
	Items     []ChoreCompletionItem
	Timer     *time.Timer
}

// Batcher accumulates chore completions and flushes them after a time window.
type Batcher struct {
	mu      sync.Mutex
	entries map[batchKey]*batchEntry
	service *Service
	stopped bool
}

// NewBatcher creates a new Batcher.
func NewBatcher(service *Service) *Batcher {
	return &Batcher{
		entries: make(map[batchKey]*batchEntry),
		service: service,
	}
}

// Start initializes the batcher. Currently a no-op since timers are per-entry.
func (b *Batcher) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopped = false
}

// Stop flushes all remaining entries and prevents new additions.
func (b *Batcher) Stop() {
	b.mu.Lock()
	b.stopped = true
	// Collect entries to flush
	toFlush := make(map[batchKey]*batchEntry)
	for k, v := range b.entries {
		if v.Timer != nil {
			v.Timer.Stop()
		}
		toFlush[k] = v
	}
	b.entries = make(map[batchKey]*batchEntry)
	b.mu.Unlock()

	// Flush outside the lock
	for k, entry := range toFlush {
		b.flush(k, entry)
	}
}

// Add queues a chore completion for batched notification.
func (b *Batcher) Add(familyID, childID int64, childName, choreName string, rewardCents int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		log.Printf("WARN: batcher: add called after stop, dropping chore completion for family %d", familyID)
		return
	}

	key := batchKey{FamilyID: familyID, ChildID: childID}
	entry, exists := b.entries[key]
	if !exists {
		entry = &batchEntry{
			ChildName: childName,
		}
		b.entries[key] = entry

		// Start timer on first add
		entry.Timer = time.AfterFunc(batchWindow, func() {
			b.mu.Lock()
			e, ok := b.entries[key]
			if ok {
				delete(b.entries, key)
			}
			b.mu.Unlock()
			if ok {
				b.flush(key, e)
			}
		})
	}

	entry.Items = append(entry.Items, ChoreCompletionItem{
		ChoreName:   choreName,
		RewardCents: rewardCents,
	})
}

// flush sends the accumulated chore completions for a batch entry.
func (b *Batcher) flush(key batchKey, entry *batchEntry) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ERROR: batcher flush panic for family %d child %d: %v", key.FamilyID, key.ChildID, r)
		}
	}()
	b.service.sendChoreCompletionEmails(key.FamilyID, entry.ChildName, entry.Items)
}
