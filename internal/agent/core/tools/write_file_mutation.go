package tools

import "sync"

type pathLock struct {
	mu       sync.Mutex
	refCount int
}

type fileMutationQueue struct {
	mu    sync.Mutex
	locks map[string]*pathLock
}

var globalMutationQueue = &fileMutationQueue{locks: make(map[string]*pathLock)}

func (q *fileMutationQueue) acquire(path string) *pathLock {
	q.mu.Lock()
	pl, ok := q.locks[path]
	if !ok {
		pl = &pathLock{}
		q.locks[path] = pl
	}
	pl.refCount++
	q.mu.Unlock()

	pl.mu.Lock()
	return pl
}

func (q *fileMutationQueue) release(path string, pl *pathLock) {
	pl.mu.Unlock()

	q.mu.Lock()
	pl.refCount--
	if pl.refCount == 0 {
		delete(q.locks, path)
	}
	q.mu.Unlock()
}

func withFileMutationQueue(path string, fn func() error) error {
	pl := globalMutationQueue.acquire(path)
	defer globalMutationQueue.release(path, pl)
	return fn()
}
