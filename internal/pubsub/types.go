// ABOUTME: Event types and interfaces for PubSub system
// ABOUTME: Defines EventType enum and Event wrapper struct

// Package pubsub provides a generic publish-subscribe event broker for decoupled communication.
package pubsub

import "context"

// EventType represents the type of event being published
type EventType int

// Event type constants for common CRUD operations
const (
	Created EventType = iota
	Updated
	Deleted
)

// Event wraps a payload with its event type
type Event[T any] struct {
	Type    EventType
	Payload T
}

// Subscriber interface for services that publish events
type Subscriber[T any] interface {
	Subscribe(ctx context.Context) <-chan Event[T]
}
