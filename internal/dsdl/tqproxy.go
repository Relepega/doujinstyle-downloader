// The package implements a Queue, a Task Database and a Wrapper to keep both in sync
//
// Queue: A basic queue implementation based on a doubly linked-lisdb.
//
// Database: A map that keeps track of the progress of every task added in idb.
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

	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db"
)

const ERR_NO_RES_FOUND = "No results found"

type (
	// function that is responsible to automatically run the queue
	QueueRunner func(tq *TQProxy, stop <-chan struct{}, opts any) error
)

// A TQProxy is a proxy that contains both Queue and Database instances.
//
// This is the recommended way of using the package with a high chance to avoid a race condition.
type TQProxy struct {
	sync.Mutex

	q  *Queue
	db db.DB

	// starter function
	qRunner QueueRunner
	// channel that should be used in the runner to stop itself
	stopRunner chan struct{}
	// whether the qRunner function is running or not
	isQueueRunning bool
	// compares every value in the DB (item) to the targt (user value)
	comparatorFn func(item, target any) bool

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
//   - dbType DBType: Type of database you want to use. If invalid, returns the default in-memory one.
//
//     To run the QueueRunner function you musk invoke the [*TQProxy.RunQueue] function
func NewTQWrapper(dbType db.DBType, fn QueueRunner, ctx context.Context) *TQProxy {
	proxy := &TQProxy{
		q:              NewQueue(),
		db:             db.GetNewDatabase(dbType),
		qRunner:        fn,
		stopRunner:     make(chan struct{}),
		isQueueRunning: false,
		comparatorFn: func(item, target any) bool {
			return item == target
		},
	}

	proxy.ctx = context.WithValue(ctx, "tq", proxy)

	return proxy
}

// same thing as NewTQWrapper, but this has to be called from dsdl engine
func newTQWrapperFromEngine(
	dbType db.DBType,
	fn QueueRunner,
	ctx context.Context,
	dsdl *DSDL,
) *TQProxy {
	proxy := &TQProxy{
		q:              NewQueue(),
		db:             db.GetNewDatabase(dbType),
		qRunner:        fn,
		stopRunner:     make(chan struct{}),
		isQueueRunning: false,
		comparatorFn: func(item, target any) bool {
			return item == target
		},
	}

	err := proxy.db.Open()
	if err != nil {
		info := fmt.Sprintf(
			"Failed to open selected db: \"%s\", falling back to the in-memory db",
			proxy.db.Name(),
		)
		fmt.Println(info)
		log.Println(info)
		proxy.db = db.GetNewDatabase(-1)
	}

	proxy.ctx = context.WithValue(ctx, "dsdl", dsdl)

	return proxy
}

// GetQueue returns the underlying pointer to the Queue instance
func (tq *TQProxy) GetQueue() *Queue {
	return tq.q
}

// GetDatabase returns the underlying pointer to the Database instance
func (tq *TQProxy) GetDatabase() db.DB {
	return tq.db
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
func (tq *TQProxy) RunQueue(opts any) {
	go func(tq *TQProxy, stop chan struct{}, opts any) {
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
func (tq *TQProxy) SetComparatorFunc(newComparator func(item, target any) bool) {
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

	alreadyExists, err := tq.db.Get(n.Value())
	if err != nil {
		return err
	}
	if alreadyExists {
		return fmt.Errorf("A node with an equal value already exists")
	}

	tq.q.Enqueue(n)
	tq.db.Add(n.Value())

	return nil
}

// Checks if the tracker already holds an equal node value.
//
// If not, creates a new Node and appends it to the queue and the tracker.
//
// Returns:
//
//   - error: returned when a Node with an equal value is found in the tracker
func (tq *TQProxy) EnqueueFromValue(value any) (any, error) {
	tq.Lock()
	defer tq.Unlock()

	tasks, err := tq.db.GetAll()
	if err != nil {
		return nil, err
	}
	for k := range tasks {
		if tq.comparatorFn(k, value) {
			return value, fmt.Errorf("A node with an equal value already exists")
		}
	}

	n := NewNode(value)

	tq.q.Enqueue(n)
	tq.db.Add(value)

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
	value any,
	comp func(item, target any) bool,
) error {
	tq.Lock()
	defer tq.Unlock()

	tasks, err := tq.db.GetAll()
	if err != nil {
		return err
	}
	for k := range tasks {
		if comp(k, value) {
			return fmt.Errorf("A node with an equal value already exists")
		}
	}

	n := NewNode(value)

	tq.q.Enqueue(n)
	tq.db.Add(value)

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
func (tq *TQProxy) Find(target any) (bool, any, error) {
	tq.Lock()
	defer tq.Unlock()

	tasks, err := tq.db.GetAll()
	if err != nil {
		return false, nil, err
	}
	for k := range tasks {
		if tq.comparatorFn(k, target) {
			return true, k, nil
		}
	}

	return false, nil, fmt.Errorf("Couldn't find a matching task")
}

// Returns all the values with the matching progress state.
//
// Returns an error if the funciton parameter is out of bounds.
func (tq *TQProxy) FindWithProgressState(state int) ([]any, error) {
	tq.Lock()
	defer tq.Unlock()

	var nodes []any

	if state < 0 || state >= db.MaxCompletionState() {
		return nodes, fmt.Errorf("State is not a value within constraints")
	}

	tasks, err := tq.db.GetAll()
	if err != nil {
		return nil, err
	}
	for k := range tasks {
		if k == state {
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
	target any,
	comp func(item, target any) bool,
) (any, error) {
	tq.Lock()
	defer tq.Unlock()

	tasks, err := tq.db.GetAll()
	if err != nil {
		return nil, err
	}
	for k := range tasks {
		if comp(k, target) {
			return k, nil
		}
	}

	return nil, fmt.Errorf("Couldn't find a matching task")
}

// Removes the node at the HEAD of the queue and returns its value
func (tq *TQProxy) Dequeue() (any, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.q.Dequeue()
}

// Removes a node from the Database by value
//
// Params:
//
//   - v any: Value of the node that will be removed
//
// Returns:
//
//   - error: Database fails to remove the node
func (tq *TQProxy) Remove(v any) error {
	tq.Lock()
	defer tq.Unlock()

	tq.q.Remove(v, func(val1, val2 any) bool {
		if val1 == val2 {
			return true
		}

		return false
	})

	err := tq.db.Remove(v)

	return err
}

// Removes a task from the Database by state
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

	if completionState < 0 || completionState >= db.MaxCompletionState() {
		return 0, fmt.Errorf("Completion state out of range")
	}

	count, err := tq.db.RemoveFromState(completionState)
	if err != nil {
		return 0, err
	}
	if count == -1 {
		return 0, fmt.Errorf("Completion state out of range")
	}

	return count, nil
}

// Removes a node from the Database by value
//
// Params:
//
//   - v any: User value
//   - comp function(v1 any, v2 any) bool: comparator function. The second param is the user value
//
// Returns:
//
//   - error: Database fails to remove the node
func (tq *TQProxy) RemoveWithComparator(
	v any,
	comp func(item, target any) bool,
) error {
	tq.Lock()
	defer tq.Unlock()

	removed, val := tq.q.Remove(v, comp)

	if val != nil && !removed {
		return fmt.Errorf("Match found but couldn't remove it")
	}

	tasks, err := tq.db.GetAll()
	if err != nil {
		return err
	}
	for k := range tasks {
		if comp(k, v) {
			tq.db.Remove(k)
		}
	}

	return nil
}

// Advances a task's state by finding it by value and incrementing its state.
//
// Returns an error when the state cannot be incremented anymore or the task cannot be found and the updated state value.
func (tq *TQProxy) AdvanceTaskState(v any) (int, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.db.AdvanceState(v)
}

// Regresses a task's state by finding it by value and decrementing its state.
//
// Returns an error when the state cannot be decremented anymore or the task cannot be found or the task is in a running state.
func (tq *TQProxy) RegressTaskState(v any) error {
	tq.Lock()
	defer tq.Unlock()

	_, err := tq.db.RegressState(v)

	return err
}

// Resets a task's state by finding it by value, resets it and appends the task as last element of the queue.
//
// Returns an error when the task cannot be found or the task is in a running state.
func (tq *TQProxy) ResetTaskState(v any) error {
	tq.Lock()
	defer tq.Unlock()

	err := tq.db.SetState(v, db.TASK_STATE_QUEUED)
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
func (tq *TQProxy) GetDatabaseCount() (int, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.db.Count()
}

// Returns the number of tasks that are in a defined completion state.
//
// It can also return an error when the completionState is invalid
func (tq *TQProxy) GetDatabaseCountFromState(completionState int) (int, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.db.CountFromState(completionState)
}

func (tq *TQProxy) Context() *DSDL {
	return tq.ctx.Value("dsdl").(*DSDL)
}
