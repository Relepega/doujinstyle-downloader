// The package implements a Queue, a Task Tracker and a Wrapper to keep both in sync
//
// Queue: A basic queue implementation based on a doubly linked-list.
//
// Tracker: A map that keeps track of the progress of every task added in it.
//
// TQWrapper: The recommended way of interacting with the package functionality if you need both queuing and tracking functionality.
//
// This wrapper ensures that everything is synchronized correctly.
package dsdl

import (
	"context"
	"fmt"
	"log"
	"sync"
)

const ERR_NO_RES_FOUND = "No results found"

type (
	// function that is responsible to automatically run the queue
	QueueRunner func(tq *TQProxy, stop <-chan struct{}, opts interface{}) error
)

// A TQProxy is a proxy that contains both Queue and Tracker instances.
//
// This is the recommended way of using the package with a high chance to avoid a race condition.
type TQProxy struct {
	sync.Mutex

	q *Queue
	t *Tracker

	// starter function
	qRunner QueueRunner
	// channel that should be used in the runner to stop itself
	stopRunner chan struct{}
	// whether the qRunner function is running or not
	isQueueRunning bool
	// compares every value in the DB (item) to the targt (user value)
	comparatorFn func(item, target interface{}) bool

	// parent context
	ctx context.Context
}

// NewTQWrapper: Returns a new pointer to TQWrapper
//
// Params:
//
//   - fn QueueRunner: function that will be run in a separate goroutine
//
//     and is responsible to automagically run the queue tasks.
//
//     To run the QueueRunner function you musk invoke the [*TQProxy.RunQueue] function
func NewTQWrapper(fn QueueRunner, ctx context.Context) *TQProxy {
	proxy := &TQProxy{
		q:              NewQueue(),
		t:              NewTracker(),
		qRunner:        fn,
		stopRunner:     make(chan struct{}),
		isQueueRunning: false,
		comparatorFn: func(item, target interface{}) bool {
			return item == target
		},
	}

	proxy.ctx = context.WithValue(ctx, "tq", proxy)

	return proxy
}

// same thing as NewTQWrapper, but this has to be called from dsdl engine
func newTQWrapperFromEngine(fn QueueRunner, ctx context.Context, dsdl *DSDL) *TQProxy {
	proxy := &TQProxy{
		q:              NewQueue(),
		t:              NewTracker(),
		qRunner:        fn,
		stopRunner:     make(chan struct{}),
		isQueueRunning: false,
		comparatorFn: func(item, target interface{}) bool {
			return item == target
		},
	}

	proxy.ctx = context.WithValue(ctx, "dsdl", dsdl)

	return proxy
}

// GetQueue returns the underlying pointer to the Queue instance
func (tq *TQProxy) GetQueue() *Queue {
	return tq.q
}

// GetTracker returns the underlying pointer to the Tracker instance
func (tq *TQProxy) GetTracker() *Tracker {
	return tq.t
}

// RunQueue: Function responsible to launch the qRunner function
//
// Params:
//
//   - opts: Generic value that holds important data
//
//     that is used to run the queue. This can be a null and has
//
//     to be casted into the proper type inside the runner fn.
func (tq *TQProxy) RunQueue(opts interface{}) {
	go func(tq *TQProxy, stop chan struct{}, opts interface{}) {
		err := tq.qRunner(tq, stop, opts)
		if err != nil {
			log.Fatalf("RunQueue: %v", err)
		}
	}(tq, tq.stopRunner, opts)

	tq.isQueueRunning = true
}

// Sends a message at the qRunner function.
//
// # The logic to stop the runner should be
//
// implemented in the function itself
func (tq *TQProxy) StopQueue() {
	tq.Lock()
	defer tq.Unlock()

	tq.stopRunner <- struct{}{}
	tq.isQueueRunning = false
}

// Returns the running status of the qRunner function
func (tq *TQProxy) IsQueueRunning() bool {
	tq.Lock()
	defer tq.Unlock()

	return tq.isQueueRunning
}

// Sets a different default comparator function.
//
// Use it if the defualt comparator function isn't working as expected
func (tq *TQProxy) SetComparatorFunc(newComparator func(item, target interface{}) bool) {
	tq.Lock()
	defer tq.Unlock()

	tq.comparatorFn = newComparator
}

// Checks if a node holding an equal value is already
//
// present in the tracker. If not, appends the node to
//
// the queue and the tracker.
//
// Returns:
//
//   - error: returned when a Node with an equal value is found in the tracker
func (tq *TQProxy) Enqueue(n *Node) error {
	tq.Lock()
	defer tq.Unlock()

	alreadyExists := tq.t.Get(n.Value())
	if alreadyExists {
		return fmt.Errorf("A node with an equal value already exists")
	}

	tq.q.Enqueue(n)
	tq.t.Add(n.Value())

	return nil
}

// Checks if the tracker already holds an equal node value.
//
// If not, creates a new Node and appends it to the queue and the tracker.
//
// Returns:
//
//   - error: returned when a Node with an equal value is found in the tracker
func (tq *TQProxy) EnqueueFromValue(value interface{}) (interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	for k := range tq.t.tasks_db {
		if tq.comparatorFn(k, value) {
			return value, fmt.Errorf("A node with an equal value already exists")
		}
	}

	n := NewNode(value)

	tq.q.Enqueue(n)
	tq.t.Add(value)

	return value, nil
}

// Checks if the tracker already holds an equal node value through a comparator function.
//
// If not, creates a new Node and appends it to the queue and the tracker.
//
// Returns:
//
//   - error: returned when a Node with an equal value is found in the tracker
func (tq *TQProxy) EnqueueFromValueWithComparator(
	value interface{},
	comp func(item, target interface{}) bool,
) error {
	tq.Lock()
	defer tq.Unlock()

	for k := range tq.t.tasks_db {
		if comp(k, value) {
			return fmt.Errorf("A node with an equal value already exists")
		}
	}

	n := NewNode(value)

	tq.q.Enqueue(n)
	tq.t.Add(value)

	return nil
}

// Checks and returns the matching task, if it exists.
//
// Returns:
//
//   - task:   task corresponding to the comparator returning a truthy value
//   - error: returned when a Node with an equal value is found in the tracker

// Finds a task by using the embedded comparator function
//
// Returns:
//
//   - found: if task is found in the db
//   - task:  task corresponding to the comparator returning a truthy value
//   - error: returned when a Node with an equal value is found in the tracker
func (tq *TQProxy) Find(target interface{}) (bool, interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	for k := range tq.t.tasks_db {
		if tq.comparatorFn(k, target) {
			return true, k, nil
		}
	}

	return false, nil, fmt.Errorf("Couldn't find a matching task")
}

// Returns all the values with the matching progress state.
//
// Returns an error if the funciton parameter is out of bounds.
func (tq *TQProxy) FindWithProgressState(state int) ([]interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	var nodes []interface{}

	if state < 0 || state >= max_completion_state {
		return nodes, fmt.Errorf("State is not a value within constraints")
	}

	for k, v := range tq.t.tasks_db {
		if v == state {
			nodes = append(nodes, k)
		}
	}

	return nodes, nil
}

// Checks and returns the matching task, if it exists, from the result of a compararion function.
//
// Returns:
//
//   - task:   task corresponding to the comparator returning a truthy value
//   - error: returned when a Node with an equal value is found in the tracker
func (tq *TQProxy) FindWithComparator(
	target interface{},
	comp func(item, target interface{}) bool,
) (interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	for k := range tq.t.tasks_db {
		if comp(k, target) {
			return k, nil
		}
	}

	return nil, fmt.Errorf("Couldn't find a matching task")
}

// Removes the node at the HEAD of the queue and returns its value
func (tq *TQProxy) Dequeue() (interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.q.Dequeue()
}

// Removes a node from the Tracker by value
//
// Params:
//
//   - v interface{}: Value of the node that will be removed
//
// Returns:
//
//   - error: Tracker fails to remove the node
func (tq *TQProxy) Remove(v interface{}) error {
	tq.Lock()
	defer tq.Unlock()

	tq.q.Remove(v, func(val1, val2 interface{}) bool {
		if val1 == val2 {
			return true
		}

		return false
	})

	err := tq.t.Remove(v)

	return err
}

// Removes a task from the Tracker by state
//
// Returns:
//
//  - the number of affected tasks.
//
//  - an error if the state is out of range

// Returns the number of affected tasks. If -1, then the state is out of range
func (tq *TQProxy) RemoveFromState(completionState int) (int, error) {
	tq.Lock()
	defer tq.Unlock()

	if completionState < 0 || completionState >= max_completion_state {
		return 0, fmt.Errorf("Completion state out of range")
	}

	count := tq.t.RemoveFromState(completionState)
	if count == -1 {
		return 0, fmt.Errorf("Completion state out of range")
	}

	return count, nil
}

// Removes a node from the Tracker by value
//
// Params:
//
//   - v interface{}: User value
//   - comp function(v1 interface{}, v2 interface{}) bool: comparator function. The second param is the user value
//
// Returns:
//
//   - error: Tracker fails to remove the node
func (tq *TQProxy) RemoveWithComparator(
	v interface{},
	comp func(item, target interface{}) bool,
) error {
	tq.Lock()
	defer tq.Unlock()

	removed, val := tq.q.Remove(v, comp)

	if val != nil && !removed {
		return fmt.Errorf("Match found but couldn't remove it")
	}

	for k := range tq.GetTracker().tasks_db {
		if comp(k, v) {
			delete(tq.GetTracker().tasks_db, k)
		}
	}

	return nil
}

// Advances a task's state by finding it by value and incrementing its state.
//
// Returns an error when the state cannot be incremented anymore or the task cannot be found and the updated state value.
func (tq *TQProxy) AdvanceTaskState(v interface{}) (int, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.t.AdvanceState(v)
}

// Regresses a task's state by finding it by value and decrementing its state.
//
// Returns an error when the state cannot be decremented anymore or the task cannot be found or the task is in a running state.
func (tq *TQProxy) RegressTaskState(v interface{}) error {
	tq.Lock()
	defer tq.Unlock()

	return tq.t.RegressState(v)
}

// Resets a task's state by finding it by value, resets it and appends the task as last element of the queue.
//
// Returns an error when the task cannot be found or the task is in a running state.
func (tq *TQProxy) ResetTaskState(v interface{}) error {
	tq.Lock()
	defer tq.Unlock()

	err := tq.t.SetState(v, TASK_STATE_QUEUED)
	if err != nil {
		return err
	}

	tq.q.Enqueue(NewNode(v))

	return nil
}

// Returns the number of tasks in the queue
func (tq *TQProxy) GetQueueLength() int {
	tq.Lock()
	defer tq.Unlock()

	return tq.q.Length()
}

// Returns the number of tasks in the tracker
func (tq *TQProxy) GetTrackerCount() int {
	tq.Lock()
	defer tq.Unlock()

	return tq.t.Count()
}

// Returns the number of tasks that are in a defined completion state.
//
// It can also return an error when the completionState is invalid
func (tq *TQProxy) GetTrackerCountFromState(completionState int) (int, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.t.CountFromState(completionState)
}

func (tq *TQProxy) Context() *DSDL {
	return tq.ctx.Value("dsdl").(*DSDL)
}
