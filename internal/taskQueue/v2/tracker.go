package queue

import (
	"fmt"
	"sync"
)

const (
	TASK_STATE_QUEUED int = iota
	TASK_STATE_RUNNING
	TASK_STATE_COMPLETED
	max_completion_state
)

const (
	TASK_STATE_STR_QUEUED    = "Queued"
	TASK_STATE_STR_RUNNING   = "Running"
	TASK_STATE_STR_COMPLETED = "Completed"
)

var statuses = map[int]string{
	TASK_STATE_QUEUED:    TASK_STATE_STR_QUEUED,
	TASK_STATE_RUNNING:   TASK_STATE_STR_RUNNING,
	TASK_STATE_COMPLETED: TASK_STATE_STR_COMPLETED,
}

type Tracker struct {
	sync.Mutex

	db_tasks map[interface{}]int
}

// TODO: MAKE A RUNNING TASK CANCELLABLE
func NewTracker() *Tracker {
	return &Tracker{
		db_tasks: make(map[interface{}]int, 15), // seems a fair, arbitrary value
	}
}

func (t *Tracker) CountFromState(completionState int) (int, error) {
	t.Lock()
	defer t.Unlock()

	if completionState < 0 || completionState >= max_completion_state {
		return -1, fmt.Errorf("Argument is not a valid state within constraints")
	}

	count := 0
	for _, v := range t.db_tasks {
		if v == completionState {
			count++
		}
	}

	return count, nil
}

func (t *Tracker) Add(nv NodeValue) {
	t.Lock()
	defer t.Unlock()

	t.db_tasks[nv] = TASK_STATE_QUEUED
}

func (t *Tracker) Has(nv NodeValue) bool {
	t.Lock()
	defer t.Unlock()

	for k := range t.db_tasks {
		if k == nv {
			return true
		}
	}
	return false
}

func (t *Tracker) Remove(nv NodeValue) error {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if k == nv {
			if v == TASK_STATE_RUNNING {
				return fmt.Errorf("Cannot remove a running task")
			}

			delete(t.db_tasks, k)
		}
	}

	return nil
}

func (t *Tracker) RemoveAll() {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if v != TASK_STATE_RUNNING {
			delete(t.db_tasks, k)
		}
	}
}

func (t *Tracker) ResetFromCompletionState(completionState int) error {
	t.Lock()
	defer t.Unlock()

	if completionState < 0 || completionState >= max_completion_state {
		return fmt.Errorf("Argument is not a valid state within constraints")
	}

	if completionState == TASK_STATE_RUNNING {
		return fmt.Errorf("Cannot cancel running tasks")
	}

	for k, v := range t.db_tasks {
		if v == completionState {
			delete(t.db_tasks, k)
		}
	}

	return nil
}

func (t *Tracker) GetStatus(nv NodeValue) (string, error) {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if k == nv {
			return statuses[v], nil
		}
	}

	return "", fmt.Errorf("Node not found")
}

func (t *Tracker) Count(completionState int) int {
	t.Lock()
	defer t.Unlock()

	count := 0

	for _, v := range t.db_tasks {
		if v == completionState {
			count++
		}
	}

	return count
}

func (t *Tracker) AdvanceState(nv NodeValue) error {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if k == nv {
			if v >= max_completion_state {
				return fmt.Errorf("Cannot advance the status of this task anymore")
			}

			t.db_tasks[k]++
		}
	}

	return nil
}

func (t *Tracker) RegressState(nv NodeValue) error {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if k == nv {
			if v <= 0 {
				return fmt.Errorf("Cannot regress the status of this task anymore")
			}

			t.db_tasks[k]--
		}
	}

	return nil
}

func (t *Tracker) ResetState(nv NodeValue) error {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if k == nv {
			if v == TASK_STATE_RUNNING {
				return fmt.Errorf("Cannot reset an already running task")
			}

			t.db_tasks[k] = TASK_STATE_QUEUED
		}
	}

	return nil
}
