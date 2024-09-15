package queue

import (
	"fmt"
	"sync"
)

type (
	QueueRunner func(tq *TQv2, stop <-chan struct{}, opts interface{})
)

type TQv2 struct {
	sync.Mutex

	q *Queue
	t *Tracker

	qRunner        QueueRunner
	stopRunner     chan struct{}
	isQueueRunning bool
}

func NewTQ(fn QueueRunner) *TQv2 {
	return &TQv2{
		q:              NewQueue(),
		t:              NewTracker(),
		qRunner:        fn,
		stopRunner:     make(chan struct{}),
		isQueueRunning: false,
	}
}

func (tq *TQv2) GetQueue() *Queue {
	return tq.q
}

func (tq *TQv2) GetTracker() *Tracker {
	return tq.t
}

func (tq *TQv2) RunQueue(opts interface{}) {
	go func(tq *TQv2, stop chan struct{}, opts interface{}) {
		tq.qRunner(tq, stop, opts)
	}(tq, tq.stopRunner, opts)

	tq.isQueueRunning = true
}

func (tq *TQv2) StopQueue() {
	tq.stopRunner <- struct{}{}
	tq.isQueueRunning = false
}

func (tq *TQv2) IsQueueRunning() bool {
	return tq.isQueueRunning
}

func (tq *TQv2) AddNode(n *Node) error {
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

func (tq *TQv2) AddNodeFromValue(v interface{}) (interface{}, error) {
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

func (tq *TQv2) Dequeue() (interface{}, error) {
	tq.Lock()
	defer tq.Unlock()

	return tq.q.Dequeue()
}

func (tq *TQv2) RemoveNode(v interface{}) error {
	tq.q.Remove(v, func(val1, val2 interface{}) bool {
		if val1 == val2 {
			return true
		}

		return false
	})

	err := tq.t.Remove(v)

	return err
}

func (tq *TQv2) AdvanceTaskState(v interface{}) error {
	return tq.t.AdvanceState(v)
}

func (tq *TQv2) AdvanceNewTaskState() (interface{}, error) {
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

func (tq *TQv2) RegressTaskState(v interface{}) error {
	return tq.t.RegressState(v)
}

func (tq *TQv2) ResetTaskState(v interface{}) error {
	return tq.t.RegressState(v)
}

func (tq *TQv2) GetQueueLength() int {
	return tq.q.Length()
}

func (tq *TQv2) TrackerCount() int {
	return tq.t.Count()
}

func (tq *TQv2) TrackerCountFromState(completionState int) (int, error) {
	return tq.t.CountFromState(completionState)
}
