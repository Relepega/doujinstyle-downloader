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
	"context"
	"fmt"
	"sync"
)

const ERR_NO_RES_FOUND = "No results found"

type (
	// function that is responsible to automatically run the queue
	QueueRunner func(tq *TQProxy, stop <-chan struct{}, opts interface{})
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
	}

	proxy.ctx = context.WithValue(ctx, "tq", proxy)

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
		tq.qRunner(tq, stop, opts)
	}(tq, tq.stopRunner, opts)

	tq.isQueueRunning = true
}

// Sends a message at the qRunner function.
//
// # The logic to stop the runner should be
//
// implemented in the function itself
func (tq *TQProxy) StopQueue() {
	tq.stopRunner <- struct{}{}
	tq.isQueueRunning = false
}

// Returns the running status of the qRunner function
func (tq *TQProxy) IsQueueRunning() bool {
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
func (tq *TQProxy) AddNode(n *Node) error {
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
func (tq *TQProxy) AddNodeFromValue(v interface{}) (interface{}, error) {
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
func (tq *TQProxy) RemoveNode(v interface{}) error {
	tq.q.Remove(v, func(val1, val2 interface{}) bool {
		if val1 == val2 {
			return true
		}

		return false
	})

	err := tq.t.Remove(v)

	return err
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
func (tq *TQProxy) RemoveNodeWithComparator(
	v interface{},
	comp func(int_v interface{}, user_v interface{}) bool,
) error {
	tq.Lock()
	defer tq.Unlock()

	removed, val := tq.q.Remove(v, comp)

	if !removed {
		return fmt.Errorf(ERR_NO_RES_FOUND)
	}

	err := tq.t.Remove(val)

	return err
}

// Returns whether or not an equal value has been found in the tracker
func (tq *TQProxy) Has(v interface{}) bool {
	tq.Lock()
	defer tq.Unlock()

	return tq.t.Has(v)
}

// Advances a task's state by finding it by value and incrementing its state.
//
// Returns an error when the state cannot be incremented anymore or the task cannot be found.
func (tq *TQProxy) AdvanceTaskState(v interface{}) error {
	return tq.t.AdvanceState(v)
}

// Advances a task's state by finding it by value and incrementing its state.
//
// Returns an error when the state cannot be incremented anymore or the task cannot be found or the task is in a running state.
func (tq *TQProxy) AdvanceNewTaskState() (interface{}, error) {
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
func (tq *TQProxy) RegressTaskState(v interface{}) error {
	return tq.t.RegressState(v)
}

// Resets a task's state by finding it by value, resets it and appends the task as last element of the queue.
//
// Returns an error when the task cannot be found or the task is in a running state.
func (tq *TQProxy) ResetTaskState(v interface{}) error {
	return tq.t.RegressState(v)
}

// Returns the number of tasks in the queue
func (tq *TQProxy) GetQueueLength() int {
	return tq.q.Length()
}

// Returns the number of tasks in the tracker
func (tq *TQProxy) TrackerCount() int {
	return tq.t.Count()
}

// Returns the number of tasks that are in a defined completion state.
//
// It can also return an error when the completionState is invalid
func (tq *TQProxy) TrackerCountFromState(completionState int) (int, error) {
	return tq.t.CountFromState(completionState)
}

func (tq *TQProxy) Context() any {
	return tq.ctx.Value("tlp")
}
