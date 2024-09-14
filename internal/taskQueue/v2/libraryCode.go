package queue

type (
	QRunnerOpts interface{}
	QueueRunner func(tq *TQv2, opts QRunnerOpts)
)

type TQv2 struct {
	q       *Queue
	t       *Tracker
	qRunner QueueRunner
}

func NewTQ(fn QueueRunner) *TQv2 {
	return &TQv2{
		q:       NewQueue(),
		t:       NewTracker(),
		qRunner: fn,
	}
}

func (tq *TQv2) GetQueue() *Queue {
	return tq.q
}

func (tq *TQv2) GetTracker() *Tracker {
	return tq.t
}

func (tq *TQv2) RunQueue(opts QRunnerOpts) {
	go func(tq *TQv2, opts QRunnerOpts) {
		tq.qRunner(tq, opts)
	}(tq, opts)
}

func (tq *TQv2) AddNode(v NodeValue) *Node {
	n := NewNode(v)

	tq.q.Enqueue(n)
	tq.t.Add(v)

	return n
}

func (tq *TQv2) RemoveNode(v NodeValue) error {
	tq.q.Remove(v, func(val1, val2 NodeValue) bool {
		if val1 == val2 {
			return true
		}

		return false
	})

	err := tq.t.Remove(v)

	return err
}

func (tq *TQv2) AdvanceTaskState(v NodeValue) error {
	return tq.t.AdvanceState(v)
}

func (tq *TQv2) RegressTaskState(v NodeValue) error {
	return tq.t.RegressState(v)
}

func (tq *TQv2) ResetTaskState(v NodeValue) error {
	return tq.t.RegressState(v)
}
