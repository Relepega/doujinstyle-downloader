// The package implements a Queue, a Task Tracker and a Wrapper to keep both in sync
//
// Queue: A basic queue implementation based on a doubly linked-list.
//
// Tracker: A map that keeps track of the progress of every task added in it.
//
// TQWrapper: The recommended way of interacting with the package functionality if you need both queuing and tracking functionality.
//
// This wrapper ensures that everything is synchronized correctly.
package queue

import (
	"fmt"
	"sync"
)

// Emun of completion states. Used to track a task's state.
const (
	TASK_STATE_QUEUED int = iota
	TASK_STATE_RUNNING
	TASK_STATE_COMPLETED
	max_completion_state
)

// String rappresentation & meaning for every completion state.
//
// The can be accessed through t.GetState()
const (
	TASK_STATE_QUEUED_STR    = "Queued"
	TASK_STATE_RUNNING_STR   = "Running"
	TASK_STATE_COMPLETED_STR = "Completed"
)

var statuses = map[int]string{
	TASK_STATE_QUEUED:    TASK_STATE_QUEUED_STR,
	TASK_STATE_RUNNING:   TASK_STATE_RUNNING_STR,
	TASK_STATE_COMPLETED: TASK_STATE_COMPLETED_STR,
}

// Tracker data type. Stores all inserted tasks in a Key-Value kind of in-memory DB
type Tracker struct {
	sync.Mutex

	db_tasks map[interface{}]int
}

// Constructor for the Tracker data type
func NewTracker() *Tracker {
	return &Tracker{
		db_tasks: make(map[interface{}]int, 15), // seems a fair, arbitrary value
	}
}

// Returns the total number of stored tasks
func (t *Tracker) Count() int {
	return len(t.db_tasks)
}

// Returns the total count of tasks in a specific completion state.
//
// Also returns an error if the specified completion state is invalid
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

// Adds a task to the Tracker
func (t *Tracker) Add(nv interface{}) {
	t.Lock()
	defer t.Unlock()

	t.db_tasks[nv] = TASK_STATE_QUEUED
}

// Checks whether a task with an equal value is already present in the Tracker
func (t *Tracker) Has(nv interface{}) bool {
	t.Lock()
	defer t.Unlock()

	for k := range t.db_tasks {
		if k == nv {
			return true
		}
	}
	return false
}

// Removes a task from the Tracker
//
// Returns an error if trying to remove a task in a running state
func (t *Tracker) Remove(nv interface{}) error {
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

// Empties the tracker
func (t *Tracker) RemoveAll() {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if v != TASK_STATE_RUNNING {
			delete(t.db_tasks, k)
		}
	}
}

// Resets the state of EVERY task in the specified completion state
//
// Returns an error either if the completion state is invalid or if trying to reset tunning tasks
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

// Returns the state of a specific task. Returns an error if the task has not been found
func (t *Tracker) GetState(nv interface{}) (string, error) {
	t.Lock()
	defer t.Unlock()

	for k, v := range t.db_tasks {
		if k == nv {
			return statuses[v], nil
		}
	}

	return "", fmt.Errorf("Node not found")
}

// Advances the completion state of a specific task
//
// Returns an error if the task has reached a completion state
func (t *Tracker) AdvanceState(nv interface{}) error {
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

// Regresses the completion state of a specific task
//
// Returns an error if the task has reached a queued state
func (t *Tracker) RegressState(nv interface{}) error {
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

// Resets the state of a specific task to a queued state
//
// Returns an error if trying to reset the state of a task in a running state
func (t *Tracker) ResetState(nv interface{}) error {
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
