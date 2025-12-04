// ABOUTME: Generic event broker for publish-subscribe pattern
// ABOUTME: Manages subscriptions and broadcasts events to all subscribers

package pubsub

import (
	"context"
	"sync"
)

// Broker manages subscriptions and event publishing for a specific type
type Broker[T any] struct {
	subscribers map[chan Event[T]]bool
	mu          sync.RWMutex
}

// NewBroker creates a new event broker
func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		subscribers: make(map[chan Event[T]]bool),
	}
}

// Subscribe creates a new subscription that receives events until context is cancelled
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan Event[T] {
	ch := make(chan Event[T], 10) // Buffer to prevent blocking publishers

	b.mu.Lock()
	b.subscribers[ch] = true
	b.mu.Unlock()

	// Cleanup when context is cancelled
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers, ch)
		close(ch)
		b.mu.Unlock()
	}()

	return ch
}

// Publish sends an event to all subscribers
func (b *Broker[T]) Publish(eventType EventType, payload T) {
	event := Event[T]{
		Type:    eventType,
		Payload: payload,
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Subscriber channel full, skip to prevent blocking
		}
	}
}
