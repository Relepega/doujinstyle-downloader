package queue

import (
	"testing"
	"time"
)

type MyDataType struct {
	value int
	state chan int
}

func runQ(tq *TQv2, opts interface{}) {
	// change interface to actual type
	// _, ok := opts.(interface{})
	// if !ok {
	// 	panic("Cannot cast runner options into proper type")
	// }

	for {
		tcount, err := tq.TrackerCountFromState(TASK_STATE_RUNNING)
		if err != nil {
			panic(err)
		}

		if tq.GetQueueLength() == 0 || tcount == 1 {
			continue
		}

		taskVal, err := tq.AdvanceNewTaskState()
		if err != nil {
			continue
		}

		v, ok := taskVal.(*MyDataType)
		if !ok {
			panic("TaskRunner: Cannot convert node value into proper type\n")
		}

		v.state <- TASK_STATE_RUNNING

		go func(t *Tracker, myData *MyDataType) {
			time.Sleep(time.Second * 5)

			err := tq.AdvanceTaskState(v)
			if err != nil {
				panic(err)
			}

			myData.state <- TASK_STATE_COMPLETED
		}(tq.GetTracker(), v)
	}
}

func TestAddNode(t *testing.T) {
	tq := NewTQ(func(tq *TQv2, opts interface{}) {})

	nv, err := tq.AddNodeFromValue(1)
	if err != nil {
		t.Fatal(err)
	}

	qlen := tq.GetQueueLength()
	tlen := tq.TrackerCount()

	if qlen != 1 {
		t.Errorf("Queue has wrong length: has %d, should be 1", qlen)
	}

	if tlen != 1 {
		t.Errorf("Tracker has wrong length: has %d, should be 1", qlen)
	}

	status, err := tq.GetTracker().GetStatus(nv)
	if err != nil {
		t.Fatal(err)
	}

	if status != TASK_STATE_STR_QUEUED {
		t.Fatalf("Wrong task status: got \"%s\", expected \"%s\"", status, TASK_STATE_STR_QUEUED)
	}
}

func TestAddRunQueue(t *testing.T) {
	tq := NewTQ(runQ)
	tq.RunQueue(nil)

	nv, err := tq.AddNodeFromValue(&MyDataType{
		value: 573,
		state: make(chan int),
	})
	if err != nil {
		t.Fatal(err)
	}

	// could be avoided, but done to check if NodeValue can be casted
	v, ok := nv.(*MyDataType)
	if !ok {
		t.Fatal("TestFN: Cannot convert Node value into proper type\n")
	}

	status, err := tq.GetTracker().GetStatus(nv)
	if err != nil {
		t.Fatal(err)
	}

	// check if status is running
	taskState := <-v.state

	if taskState != TASK_STATE_RUNNING {
		t.Fatalf("Wrong task status: got \"%d\", expected \"%d\"", taskState, TASK_STATE_RUNNING)
	}

	tcount, err := tq.TrackerCountFromState(TASK_STATE_RUNNING)
	if err != nil {
		t.Fatal(err)
	}

	if tcount != 1 {
		t.Errorf("Expected 1 task to be running, but got %d", tcount)
	}

	qlen := tq.GetQueueLength()

	if qlen != 0 {
		t.Errorf("Queue has wrong length: has %d, should be 0", qlen)
	}

	status, err = tq.GetTracker().GetStatus(nv)
	if err != nil {
		t.Fatal(err)
	}

	if status != TASK_STATE_STR_RUNNING {
		t.Fatalf("Wrong task status: got \"%s\", expected \"%s\"", status, TASK_STATE_STR_RUNNING)
	}

	// queue length & tracker count after the task should be done
	taskState = <-v.state

	qlen = tq.GetQueueLength()
	tlen := tq.TrackerCount()

	if qlen != 0 {
		t.Errorf("Queue has wrong length: has %d, should be 0", qlen)
	}

	if tlen != 1 {
		t.Errorf("Tracker has wrong length: has %d, should be 1", qlen)
	}

	// check if status is completed
	if taskState != TASK_STATE_COMPLETED {
		t.Fatalf("Wrong task status: got \"%d\", expected \"%d\"", taskState, TASK_STATE_COMPLETED)
	}

	status, err = tq.GetTracker().GetStatus(nv)
	if err != nil {
		t.Fatal(err)
	}

	if status != TASK_STATE_STR_COMPLETED {
		t.Fatalf("Wrong task status: got \"%s\", expected \"%s\"", status, TASK_STATE_STR_COMPLETED)
	}
}
