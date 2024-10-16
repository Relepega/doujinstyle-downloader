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

type (
	// function that is responsible to automatically run the queue
	QueueRunner func(tq *TQWrapper, stop <-chan struct{}, opts interface{})
)

// A TQWrapper is a wrapper that contains both Queue and Tracker instances.
//
// This is the recommended way of using the package with a high chance to avoid a race condition.
type TQWrapper struct {
	sync.Mutex

	q *Queue
	t *Tracker

	// starter function
	qRunner QueueRunner
	// channel that should be used in the runner to stop itself
	stopRunner chan struct{}
	// whether the qRunner function is running or not
	isQueueRunning bool
}

// NewTQWrapper: Returns a new pointer to TQWrapper
//
// Params:
//
//   - fn QueueRunner: function that will be run in a separate goroutine
//
//     and is responsible to automagically run the queue tasks.
//
//     To run the QueueRunner function you musk invoke the [*TQWrapper.RunQueue] function
func NewTQWrapper(fn QueueRunner) *TQWrapper {
	return &TQWrapper{
		q:              NewQueue(),
		t:              NewTracker(),
		qRunner:        fn,
		stopRunner:     make(chan struct{}),
		isQueueRunning: false,
	}
}

// GetQueue returns the underlying pointer to the Queue instance
func (tq *TQWrapper) GetQueue() *Queue {
	return tq.q
}

// GetTracker returns the underlying pointer to the Tracker instance
func (tq *TQWrapper) GetTracker() *Tracker {
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
func (tq *TQWrapper) RunQueue(opts interface{}) {
	go func(tq *TQWrapper, stop chan struct{}, opts interface{}) {
		tq.qRunner(tq, stop, opts)
	}(tq, tq.stopRunner, opts)

	tq.isQueueRunning = true
}

// Sends a message at the qRunner function.
//
// # The logic to stop the runner should be
//
// implemented in the function itself
func (tq *TQWrapper) StopQueue() {
	tq.stopRunner <- struct{}{}
	tq.isQueueRunning = false
}

// Returns the running status of the qRunner function
func (tq *TQWrapper) IsQueueRunning() bool {
	return tq.isQueueRunning
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
func (tq *TQWrapper) AddNode(n *Node) error {
	tq.Lock()
	defer tq.Unlock()

	alreadyExists := tq.t.Has(n.Value())
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
func (tq *TQWrapper) AddNodeFromValue(v interface{}) (interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	alreadyExists := tq.t.Has(v)
	if alreadyExists {
		return nil, fmt.Errorf("A node with an equal value already exists")
	}

	n := NewNode(v)

	tq.q.Enqueue(n)
	tq.t.Add(v)

	return n.Value(), nil
}

// Removes the node at the HEAD of the queue and returns its value
func (tq *TQWrapper) Dequeue() (interface{}, error) {
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
func (tq *TQWrapper) RemoveNode(v interface{}) error {
	tq.q.Remove(v, func(val1, val2 interface{}) bool {
		if val1 == val2 {
			return true
		}

		return false
	})

	err := tq.t.Remove(v)

	return err
}

// Returns whether or not an equal value has been found in the tracker
func (tq *TQWrapper) Has(v interface{}) bool {
	tq.Lock()
	defer tq.Unlock()

	return tq.t.Has(v)
}

// Advances a task's state by finding it by value and incrementing its state.
//
// Returns an error when the state cannot be incremented anymore or the task cannot be found.
func (tq *TQWrapper) AdvanceTaskState(v interface{}) error {
	return tq.t.AdvanceState(v)
}

// Advances a task's state by finding it by value and incrementing its state.
//
// Returns an error when the state cannot be incremented anymore or the task cannot be found or the task is in a running state.
func (tq *TQWrapper) AdvanceNewTaskState() (interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	nv, err := tq.q.Dequeue()
	if err != nil {
		return nil, err
	}

	err = tq.t.AdvanceState(nv)
	if err != nil {
		return nil, err
	}

	return nv, nil
}

// Regresses a task's state by finding it by value and decrementing its state.
//
// Returns an error when the state cannot be decremented anymore or the task cannot be found or the task is in a running state.
func (tq *TQWrapper) RegressTaskState(v interface{}) error {
	return tq.t.RegressState(v)
}

// Resets a task's state by finding it by value, resets it and appends the task as last element of the queue.
//
// Returns an error when the task cannot be found or the task is in a running state.
func (tq *TQWrapper) ResetTaskState(v interface{}) error {
	return tq.t.RegressState(v)
}

// Returns the number of tasks in the queue
func (tq *TQWrapper) GetQueueLength() int {
	return tq.q.Length()
}

// Returns the number of tasks in the tracker
func (tq *TQWrapper) TrackerCount() int {
	return tq.t.Count()
}

// Returns the number of tasks that are in a defined completion state.
//
// It can also return an error when the completionState is invalid
func (tq *TQWrapper) TrackerCountFromState(completionState int) (int, error) {
	return tq.t.CountFromState(completionState)
}
