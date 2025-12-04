package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBroker_PublishSubscribe(t *testing.T) {
	t.Parallel()
	broker := NewBroker[string]()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	events := broker.Subscribe(ctx)

	broker.Publish(Created, "test-payload")

	select {
	case event := <-events:
		assert.Equal(t, Created, event.Type)
		assert.Equal(t, "test-payload", event.Payload)
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	t.Parallel()
	broker := NewBroker[int]()

	ctx := context.Background()
	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)

	broker.Publish(Created, 42)

	val1 := <-sub1
	val2 := <-sub2

	assert.Equal(t, 42, val1.Payload)
	assert.Equal(t, 42, val2.Payload)
}

func TestBroker_Unsubscribe(t *testing.T) {
	t.Parallel()
	broker := NewBroker[string]()

	ctx, cancel := context.WithCancel(context.Background())
	events := broker.Subscribe(ctx)

	cancel() // Unsubscribe
	time.Sleep(50 * time.Millisecond)

	broker.Publish(Created, "test")

	select {
	case event, ok := <-events:
		if ok {
			t.Fatalf("received event after unsubscribe: %+v", event)
		}
		// Channel closed - this is expected
	case <-time.After(50 * time.Millisecond):
		// Timeout is also acceptable - no event received
	}
}
