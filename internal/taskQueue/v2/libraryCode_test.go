package queue

import (
	"fmt"
	"testing"
	"time"
)

type TestingDataType struct {
	value int
	err   error
	state chan int
}

type TestingRunnerOptions struct {
	MaxConcurrency int
	TaskDuration   time.Duration
}

func NewTestingRunnerOpts(c int, d time.Duration) TestingRunnerOptions {
	return TestingRunnerOptions{
		MaxConcurrency: c,
		TaskDuration:   d,
	}
}

func runQ(tq *TQv2, opts interface{}) {
	options := TestingRunnerOptions{
		MaxConcurrency: 1,
		TaskDuration:   time.Second,
	}

	if opts != nil {
		fnOpts, ok := opts.(TestingRunnerOptions)
		if !ok {
			panic("Cannot cast runner options into proper type")
		}

		options = fnOpts
	}

	for {
		tcount, err := tq.TrackerCountFromState(TASK_STATE_RUNNING)
		if err != nil {
			continue
		}

		if tq.GetQueueLength() == 0 || tcount == options.MaxConcurrency {
			time.Sleep(time.Millisecond)
			continue
		}

		taskVal, err := tq.AdvanceNewTaskState()
		if err != nil {
			continue
		}

		v, ok := taskVal.(*TestingDataType)
		if !ok {
			panic("TaskRunner: Cannot convert node value into proper type\n")
		}
		v.state <- TASK_STATE_RUNNING

		go func(t *Tracker, myData *TestingDataType, duration time.Duration) {
			fmt.Println("activating task")

			time.Sleep(duration)

			err = tq.AdvanceTaskState(myData)
			if err != nil {
				panic(err)
			}
			myData.state <- TASK_STATE_COMPLETED

			fmt.Println("task done")
		}(tq.GetTracker(), v, options.TaskDuration)
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

func TestRunQueue(t *testing.T) {
	tq := NewTQ(runQ)
	tq.RunQueue(NewTestingRunnerOpts(1, time.Second*2))

	nv, err := tq.AddNodeFromValue(&TestingDataType{
		value: 573,
		err:   nil,
		state: make(chan int, 1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// could be avoided, but done to check if NodeValue can be casted
	v, ok := nv.(*TestingDataType)
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
		t.Errorf("Tracker has wrong length: has %d, should be 1", tlen)
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

func TestMultipleCoroutines(t *testing.T) {
	fmt.Println("--------------------")
	tq := NewTQ(runQ)
	tq.RunQueue(NewTestingRunnerOpts(4, time.Second*5))

	ntasks := 1000

	for i := 0; i < 1000; i++ {
		_, err := tq.AddNodeFromValue(&TestingDataType{
			value: i,
			err:   nil,
			state: make(chan int, 2), // make it buffered so that the runner goroutine isn't blocked
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
