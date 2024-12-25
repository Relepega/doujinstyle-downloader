package dsdl

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type testingDataType struct {
	value int
	err   error

	state chan int
	stop  chan struct{}
}

type testingRunnerOptions struct {
	Threads      int
	TaskDuration time.Duration
}

func newTestingRunnerOpts(t int, d time.Duration) testingRunnerOptions {
	return testingRunnerOptions{
		Threads:      t,
		TaskDuration: d,
	}
}

func runQ(tq *TQProxy, stop <-chan struct{}, opts interface{}) error {
	options := testingRunnerOptions{
		Threads:      1,
		TaskDuration: time.Second,
	}

	if opts != nil {
		fnOpts, ok := opts.(testingRunnerOptions)
		if !ok {
			panic("Cannot cast runner options into proper type")
		}

		options = fnOpts
	}

	for {
		select {
		case <-stop:
			return nil

		default:
			tcount, err := tq.TrackerCountFromState(TASK_STATE_RUNNING)
			if err != nil {
				continue
			}

			if tq.GetQueueLength() == 0 || tcount == options.Threads {
				time.Sleep(time.Millisecond)
				continue
			}

			taskVal, _, err := tq.AdvanceNewTaskState()
			if err != nil {
				continue
			}

			v, ok := taskVal.(*testingDataType)
			if !ok {
				panic("TaskRunner: Cannot convert node value into proper type\n")
			}
			v.state <- TASK_STATE_RUNNING

			go taskRunner(tq, v, options.TaskDuration)
		}
	}
}

func taskRunner(tq *TQProxy, myData *testingDataType, duration time.Duration) {
	markCompleted := func() {
		_, err := tq.AdvanceTaskState(myData)
		if err != nil {
			panic(err)
		}
		myData.state <- TASK_STATE_COMPLETED
	}

	running := false

	for {
		select {
		case <-myData.stop:
			myData.err = fmt.Errorf("task aborted by the user")
			markCompleted()

			return

		default:
			if running {
				continue
			}

			// mark running
			running = true

			// intensive task operations...
			time.Sleep(duration)

			// task done :)
			markCompleted()
		}
	}
}

func TestAddNode(t *testing.T) {
	tq := NewTQWrapper(
		func(tq *TQProxy, stop <-chan struct{}, opts interface{}) error { return nil },
		context.Background(),
	)

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

	status, err := tq.GetTracker().GetState(nv)
	if err != nil {
		t.Fatal(err)
	}

	if status != TASK_STATE_QUEUED_STR {
		t.Fatalf("Wrong task status: got \"%s\", expected \"%s\"", status, TASK_STATE_QUEUED_STR)
	}
}

func TestHasNode(t *testing.T) {
	tq := NewTQWrapper(
		func(tq *TQProxy, stop <-chan struct{}, opts interface{}) error { return nil },
		context.Background(),
	)

	node := NewNode(1)

	err := tq.AddNode(node)
	if err != nil {
		t.Fatal(err)
	}

	if !tq.Has(node.value) {
		t.Errorf("TQ should already hold this value: %+v", node.value)
	}
}

func TestRunQueue(t *testing.T) {
	tq := NewTQWrapper(runQ, context.Background())
	tq.RunQueue(newTestingRunnerOpts(1, time.Second*2))

	nv, err := tq.AddNodeFromValue(&testingDataType{
		value: 573,
		err:   nil,
		state: make(chan int, 1),
		stop:  make(chan struct{}),
	})
	if err != nil {
		t.Fatal(err)
	}

	// could be avoided, but done to check if NodeValue can be casted
	v, ok := nv.(*testingDataType)
	if !ok {
		t.Fatal("TestFN: Cannot convert Node value into proper type\n")
	}

	status, err := tq.GetTracker().GetState(nv)
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

	status, err = tq.GetTracker().GetState(nv)
	if err != nil {
		t.Fatal(err)
	}

	if status != TASK_STATE_RUNNING_STR {
		t.Fatalf("Wrong task status: got \"%s\", expected \"%s\"", status, TASK_STATE_RUNNING_STR)
	}

	// queue length & tracker count after the task should be done
	taskState = <-v.state

	qlen = tq.GetQueueLength()
	tlen := tq.TrackerCount()

	if qlen != 0 {
		t.Errorf("Queue has wrong length: has %d, should be 0", qlen)
	}

	if tlen != 1 {
		t.Errorf("Tracker has wrong length: has %d, should be 1", tlen)
	}

	// check if status is completed
	if taskState != TASK_STATE_COMPLETED {
		t.Fatalf("Wrong task status: got \"%d\", expected \"%d\"", taskState, TASK_STATE_COMPLETED)
	}

	status, err = tq.GetTracker().GetState(nv)
	if err != nil {
		t.Fatal(err)
	}

	if status != TASK_STATE_COMPLETED_STR {
		t.Fatalf("Wrong task status: got \"%s\", expected \"%s\"", status, TASK_STATE_COMPLETED_STR)
	}
}

func TestMultipleCoroutines(t *testing.T) {
	tq := NewTQWrapper(runQ, context.Background())
	tq.RunQueue(newTestingRunnerOpts(4, time.Second*5))

	ntasks := 1000

	for i := 0; i < ntasks; i++ {
		_, err := tq.AddNodeFromValue(&testingDataType{
			value: i,
			err:   nil,
			// make it buffered so that the runner goroutine isn't blocked
			state: make(chan int, 1),
			stop:  make(chan struct{}),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 2)

	qlen := tq.GetQueueLength()
	tlen := tq.TrackerCount()

	if qlen != ntasks-4 {
		t.Errorf("Queue has wrong length: has %d, should be %d", qlen, ntasks-4)
	}

	if tlen != ntasks {
		t.Errorf("Tracker has wrong length: has %d, should be %d", tlen, ntasks)
	}

	count, err := tq.TrackerCountFromState(TASK_STATE_RUNNING)
	if err != nil {
		t.Fatal(err)
	}

	if count != 4 {
		t.Errorf("Running tasks should be %d, instead got %d", 4, count)
	}
}

func TestAbortTask(t *testing.T) {
	tq := NewTQWrapper(runQ, context.Background())
	tq.RunQueue(newTestingRunnerOpts(4, time.Second*5))

	nv1 := &testingDataType{
		value: 420,
		err:   nil,
		// make it buffered so that the runner goroutine isn't blocked
		state: make(chan int, 1),
		stop:  make(chan struct{}),
	}

	nv2 := &testingDataType{
		value: 727,
		err:   nil,
		// make it buffered so that the runner goroutine isn't blocked
		state: make(chan int, 1),
		stop:  make(chan struct{}),
	}

	_, err := tq.AddNodeFromValue(nv1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tq.AddNodeFromValue(nv2)
	if err != nil {
		t.Fatal(err)
	}

	// wait for both to be running
	<-nv1.state
	<-nv2.state

	nv1.stop <- struct{}{}

	// should be completed
	_ = <-nv1.state

	tq.StopQueue()

	countDone, err := tq.TrackerCountFromState(TASK_STATE_COMPLETED)
	if err != nil {
		t.Fatal(err)
	}

	if countDone != 1 {
		t.Error("Wrong done count number: should be 1, got ", countDone)
	}
}

func TestCloseRunner(t *testing.T) {
	tq := NewTQWrapper(runQ, context.Background())
	tq.RunQueue(newTestingRunnerOpts(4, time.Second*5))

	ntasks := 1000

	for i := 0; i < ntasks; i++ {
		_, err := tq.AddNodeFromValue(&testingDataType{
			value: i,
			err:   nil,
			// make it buffered so that the runner goroutine isn't blocked
			state: make(chan int, 1),
			stop:  make(chan struct{}),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 2)

	isRunning := tq.IsQueueRunning()
	if !isRunning {
		t.Fatal("QRunner should be running")
	}

	tq.StopQueue()

	isRunning = tq.IsQueueRunning()
	if isRunning {
		t.Fatal("QRunner should be stopped")
	}
}
