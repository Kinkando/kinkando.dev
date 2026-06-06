package event

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// Type identifies a domain event.
type Type string

const (
	MedicineTaken          Type = "medicine.taken"
	SupplementTaken        Type = "supplement.taken"
	WorkoutSessionFinished Type = "workout.session.finished"
	WeightLogged           Type = "weight.logged"
	SleepLogged            Type = "sleep.logged"
	QuestCompleted         Type = "quest.completed"
)

// Event carries the event type and the user it belongs to.
type Event struct {
	Type   Type
	UserID uuid.UUID
}

// Handler is a function that reacts to an event.
type Handler func(ctx context.Context, e Event)

// Bus is a minimal synchronous, in-process pub/sub.
// Publishers and subscribers depend only on this package — they never import each other.
type Bus struct {
	mu       sync.RWMutex
	handlers map[Type][]Handler
}

// New returns a ready-to-use Bus.
func New() *Bus {
	return &Bus{handlers: make(map[Type][]Handler)}
}

// Subscribe registers h to be called whenever an event of type t is published.
// Safe to call from any goroutine.
func (b *Bus) Subscribe(t Type, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[t] = append(b.handlers[t], h)
}

// Publish delivers e to every registered handler for e.Type.
// Handlers run synchronously in registration order.
// Each handler is wrapped in a recover so a panicking subscriber never
// breaks the calling (publishing) request.
func (b *Bus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	hs := b.handlers[e.Type]
	b.mu.RUnlock()

	for _, h := range hs {
		func() {
			defer func() { recover() }() //nolint:errcheck
			h(ctx, e)
		}()
	}
}
