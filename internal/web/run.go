package web

import (
	"context"
	"errors"
	"sync"
)

const (
	maxRunBufferBytes = 4 << 20
	maxSubscribers    = 16
	subscriberBuffer  = 256
)

var (
	ErrRunActive     = errors.New("a run is already active")
	ErrRunBufferFull = errors.New("run replay buffer is full")
	ErrTooManySubs   = errors.New("too many subscribers")
)

type Registry struct {
	ctx    context.Context
	mu     sync.Mutex
	active *Run
	wg     sync.WaitGroup
}

func NewRegistry(ctx context.Context) *Registry {
	return &Registry{ctx: ctx}
}

func (r *Registry) Start(conversationID int64) (*Run, *Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active != nil {
		return nil, nil, ErrRunActive
	}
	ctx, cancel := context.WithCancel(r.ctx)
	run := &Run{
		ConversationID: conversationID,
		ctx:            ctx,
		cancel:         cancel,
		registry:       r,
		subs:           make(map[*Subscription]struct{}),
	}
	sub, err := run.subscribeLocked()
	if err != nil {
		cancel()
		return nil, nil, err
	}
	r.active = run
	return run, sub, nil
}

func (r *Registry) Active() *Run {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active
}

func (r *Registry) Subscribe(conversationID int64) (*Subscription, error) {
	run := r.Active()
	if run == nil || run.ConversationID != conversationID {
		return nil, nil
	}
	run.mu.Lock()
	defer run.mu.Unlock()
	if run.done {
		return nil, nil
	}
	return run.subscribeLocked()
}

func (r *Registry) Stop(conversationID int64) {
	run := r.Active()
	if run == nil || run.ConversationID != conversationID {
		return
	}
	run.stop()
}

func (r *Registry) Shutdown(ctx context.Context) {
	r.StopAll()
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (r *Registry) StopAll() {
	run := r.Active()
	if run != nil {
		run.stop()
	}
}

func (r *Registry) clear(run *Run) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active == run {
		r.active = nil
	}
}

type Run struct {
	ConversationID int64

	ctx      context.Context
	cancel   context.CancelFunc
	registry *Registry

	mu        sync.Mutex
	frames    [][]byte
	size      int
	subs      map[*Subscription]struct{}
	done      bool
	truncated bool
	stopped   bool
}

func (r *Run) Context() context.Context {
	return r.ctx
}

func (r *Run) Emit(frame []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return
	}
	if !r.truncated {
		if r.size+len(frame) > maxRunBufferBytes {
			r.truncated = true
		} else {
			r.frames = append(r.frames, frame)
			r.size += len(frame)
		}
	}
	for sub := range r.subs {
		select {
		case sub.ch <- frame:
		default:
			delete(r.subs, sub)
			close(sub.ch)
		}
	}
}

func (r *Run) Execute(fn func(context.Context, *Run) error) {
	r.registry.wg.Add(1)
	defer r.registry.wg.Done()

	err := fn(r.ctx, r)
	if err != nil {
		if r.isStopped() {
			r.Emit(fAbort("user cancelled"))
		} else {
			r.Emit(fError(err.Error()))
		}
	}
	r.Emit(fFinish())
	r.Emit([]byte(frameDone))
	r.finish()
}

func (r *Run) stop() {
	r.mu.Lock()
	r.stopped = true
	r.mu.Unlock()
	r.cancel()
}

func (r *Run) isStopped() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopped
}

func (r *Run) finish() {
	r.mu.Lock()
	r.done = true
	for sub := range r.subs {
		delete(r.subs, sub)
		close(sub.ch)
	}
	r.mu.Unlock()
	r.cancel()
	r.registry.clear(r)
}

func (r *Run) subscribeLocked() (*Subscription, error) {
	if r.truncated {
		return nil, ErrRunBufferFull
	}
	if len(r.subs) >= maxSubscribers {
		return nil, ErrTooManySubs
	}
	sub := &Subscription{
		ch:     make(chan []byte, subscriberBuffer),
		replay: append([][]byte(nil), r.frames...),
		run:    r,
	}
	r.subs[sub] = struct{}{}
	return sub, nil
}

func (r *Run) unsubscribe(sub *Subscription) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.subs[sub]; ok {
		delete(r.subs, sub)
		close(sub.ch)
	}
}

type Subscription struct {
	ch     chan []byte
	replay [][]byte
	run    *Run

	closeOnce sync.Once
}

func (s *Subscription) Replay() [][]byte      { return s.replay }
func (s *Subscription) Frames() <-chan []byte { return s.ch }

func (s *Subscription) Close() {
	s.closeOnce.Do(func() {
		s.run.unsubscribe(s)
	})
}
