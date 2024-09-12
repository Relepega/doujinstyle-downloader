package queue

import "sync"

type Tracker[T any] struct {
	mu sync.Mutex

	queued    []*Node[T]
	running   []*Node[T]
	completed []*Node[T]
}

func (t *Tracker[T]) CountQueued() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.queued)
}

func (t *Tracker[T]) CountRunning() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.running)
}

func (t *Tracker[T]) CountCompleted() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.completed)
}

func (t *Tracker[T]) AddQueued(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.queued = append(t.queued, n)
}

func (t *Tracker[T]) AddRunning(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.running = append(t.running, n)
}

func (t *Tracker[T]) AddCompleted(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.completed = append(t.completed, n)
}

func (t *Tracker[T]) RemoveQueued(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, tn := range t.queued {
		if tn == n {
			t.queued = append(t.queued[:i], t.queued[i+1:]...)
		}
	}
}

func (t *Tracker[T]) RemoveRunning(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, tn := range t.queued {
		if tn == n {
			t.running = append(t.running[:i], t.running[i+1:]...)
		}
	}
}

func (t *Tracker[T]) RemoveCompleted(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, tn := range t.queued {
		if tn == n {
			t.completed = append(t.completed[:i], t.completed[i+1:]...)
		}
	}
}

func (t *Tracker[T]) ResetQueued(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.queued = make([]*Node[T], 0)
}

func (t *Tracker[T]) ResetRunning(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.running = make([]*Node[T], 0)
}

func (t *Tracker[T]) ResetCompleted(n *Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.completed = make([]*Node[T], 0)
}
