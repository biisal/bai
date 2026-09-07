package tui

import "testing"

func TestMessageQueueDequeueFIFO(t *testing.T) {
	var q messageQueue

	q.Enqueue("first")
	q.Enqueue("second")
	q.Enqueue("third")

	if q.Len() != 3 {
		t.Fatalf("expected len 3, got %d", q.Len())
	}

	tests := []string{"first", "second", "third"}
	for _, want := range tests {
		got, ok := q.Dequeue()
		if !ok {
			t.Fatalf("expected queued message %q", want)
		}
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}

	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got len %d", q.Len())
	}

	if got, ok := q.Dequeue(); ok || got != "" {
		t.Fatalf("expected empty dequeue, got %q, %v", got, ok)
	}
}
