package broker

import (
	"context"
	"testing"
	"time"
)

func TestPublishDoesNotBlockWhenDrained(t *testing.T) {
	b := New()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events := b.Subscribe()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-events:
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			b.Publish(ctx, Message{Type: EventAgentResponse, Text: "event"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("publisher blocked with a drain running")
	}
}
