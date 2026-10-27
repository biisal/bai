package web

import (
	"context"
	"errors"
	"testing"

	test_utils "github.com/biisal/bai/utils/tests"
)

func collect(t *testing.T, sub *Subscription) []string {
	t.Helper()
	out := make([]string, 0)
	for _, frame := range sub.Replay() {
		out = append(out, string(frame))
	}
	for frame := range sub.Frames() {
		out = append(out, string(frame))
	}
	return out
}

func TestRunTwoSubscribersSameFrames(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, first, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := registry.Subscribe(1)
	if err != nil {
		t.Fatal(err)
	}

	run.Emit([]byte("a"))
	run.Emit([]byte("b"))
	run.finish()

	test_utils.AssertSliceEqual(t, collect(t, first), []string{"a", "b"})
	test_utils.AssertSliceEqual(t, collect(t, second), []string{"a", "b"})
}

func TestRunLateSubscriberReplays(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, _, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	run.Emit([]byte("a"))
	run.Emit([]byte("b"))

	late, err := registry.Subscribe(1)
	if err != nil {
		t.Fatal(err)
	}
	run.Emit([]byte("c"))
	run.finish()

	test_utils.AssertSliceEqual(t, collect(t, late), []string{"a", "b", "c"})
}

func TestRunDropsStalledSubscriberWithoutBlocking(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, stalled, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}

	total := subscriberBuffer + 50
	for i := 0; i < total; i++ {
		run.Emit([]byte("frame"))
	}

	late, err := registry.Subscribe(1)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(late.Replay()); got != total {
		t.Fatalf("late replay length = %d, want %d", got, total)
	}

	if got := len(collect(t, stalled)); got != subscriberBuffer {
		t.Fatalf("stalled subscriber received %d frames, want %d", got, subscriberBuffer)
	}

	run.finish()
}

func TestRunSecondStartRefused(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, _, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := registry.Start(2); !errors.Is(err, ErrRunActive) {
		t.Fatalf("got %v, want %v", err, ErrRunActive)
	}
	run.finish()
	if _, _, err := registry.Start(2); err != nil {
		t.Fatalf("start after finish: %v", err)
	}
}

func TestRunExecuteFinishesAndClears(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, sub, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		run.Execute(func(_ context.Context, r *Run) error {
			r.Emit([]byte("x"))
			return nil
		})
		close(done)
	}()
	<-done

	if registry.Active() != nil {
		t.Fatal("registry still has an active run")
	}
	test_utils.AssertSliceEqual(t, collect(t, sub), []string{"x", `{"type":"finish"}`, "[DONE]"})
}

func TestRunErrorAndAbortFrames(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, sub, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	run.Execute(func(context.Context, *Run) error { return errors.New("boom") })
	got := collect(t, sub)
	want := []string{`{"type":"error","errorText":"boom"}`, `{"type":"finish"}`, "[DONE]"}
	test_utils.AssertSliceEqual(t, got, want)

	run2, sub2, err := registry.Start(2)
	if err != nil {
		t.Fatal(err)
	}
	run2.stop()
	run2.Execute(func(context.Context, *Run) error { return context.Canceled })
	got = collect(t, sub2)
	want = []string{`{"type":"abort","reason":"user cancelled"}`, `{"type":"finish"}`, "[DONE]"}
	test_utils.AssertSliceEqual(t, got, want)
}

func TestSubscribeOnlyActiveConversation(t *testing.T) {
	registry := NewRegistry(context.Background())
	run, _, err := registry.Start(1)
	if err != nil {
		t.Fatal(err)
	}
	if sub, err := registry.Subscribe(2); err != nil || sub != nil {
		t.Fatalf("got sub=%v err=%v, want nil, nil", sub, err)
	}
	run.finish()
	if sub, err := registry.Subscribe(1); err != nil || sub != nil {
		t.Fatalf("got sub=%v err=%v, want nil, nil after finish", sub, err)
	}
}
