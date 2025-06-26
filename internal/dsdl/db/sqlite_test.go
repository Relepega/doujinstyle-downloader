package db

import (
	"strings"
	"testing"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

func TestCreateDB(t *testing.T) {
	db := GetNewDatabase[*task.Task](DB_SQlite)

	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInsertAndCount(t *testing.T) {
	db := GetNewDatabase[*task.Task](DB_SQlite)

	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = db.Insert(task.NewTask("hello"))
	if err != nil {
		t.Fatal(err)
	}

	db.Insert(task.NewTask("world"))
	if err != nil {
		t.Fatal(err)
	}

	db.Insert(task.NewTask("sqlite"))
	if err != nil {
		t.Fatal(err)
	}

	count, err := db.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal("Expected count: 3, returned:", count)
	}

	err = db.Drop("*")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	db := GetNewDatabase[*task.Task](DB_SQlite)

	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = db.Insert(task.NewTask("hello"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Insert(task.NewTask("world"))
	if err != nil {
		t.Fatal(err)
	}

	t1 := task.NewTask("sqlite")
	_, err = db.Insert(t1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Insert(task.NewTask("sqlite"))
	if err != nil {
		t.Fatal(err)
	}

	partialSlug := "ite"

	tsk, err := db.Get(partialSlug)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains((*tsk).GetSlug(), partialSlug) {
		t.Fatal("Wrong task found")
	}

	tasks, err := db.GetAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 4 {
		t.Fatalf("Getall size mismatch: wanted 4 entries, got %d", len(tasks))
	}

	if tasks[2].GetID() != t1.GetID() {
		t.Fatalf("IDs mismatch: wanted: %v, got %v", tasks[2].GetID(), t1.GetID())
	}

	err = db.Drop("*")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	db := GetNewDatabase[*task.Task](DB_SQlite)

	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Insert(task.NewTask("hello"))
	if err != nil {
		t.Fatal(err)
	}

	t1 := task.NewTask("world")
	_, err = db.Insert(t1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Insert(task.NewTask("sqlite"))
	if err != nil {
		t.Fatal(err)
	}

	t2 := task.NewTask("apples")
	t2.DownloadState = states.TASK_STATE_RUNNING

	_, err = db.Insert(t2)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Remove(t1)
	if err != nil {
		t.Fatal(err)
	}

	count, err := db.Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 3 {
		t.Fatalf("DB:Remove: Expected 3 records left, got %d", count)
	}

	delCount, err := db.RemoveFromState(states.TASK_STATE_RUNNING)
	if err != nil {
		t.Fatal(err)
	}

	if delCount != 1 {
		t.Fatalf("DB:RemoveFromState: expected 1 deletion, got %d", count)
	}

	count, err = db.Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatalf("DB:Remove: Expected 2 records left, got %d", count)
	}

	err = db.RemoveAll()
	if err != nil {
		t.Fatal(err)
	}

	count, err = db.Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("DB:RemoveAll: Expected 0 records left, got %d", count)
	}

	err = db.Drop("*")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStateManipulation1(t *testing.T) {
	db := GetNewDatabase[*task.Task](DB_SQlite)

	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Insert(task.NewTask("hello"))
	if err != nil {
		t.Fatal(err)
	}

	t1 := task.NewTask("world")
	t1.DownloadState = states.TASK_STATE_COMPLETED

	_, err = db.Insert(t1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Insert(task.NewTask("sqlite"))
	if err != nil {
		t.Fatal(err)
	}

	t2 := task.NewTask("apples")

	_, err = db.Insert(t2)
	if err != nil {
		t.Fatal(err)
	}

	err = db.SetState(t2, states.TASK_STATE_COMPLETED)
	if err != nil {
		t.Fatal(err)
	}
	t2.DownloadState = states.TASK_STATE_COMPLETED

	state, err := db.GetState(t2)
	if state != states.GetStateStr(t2.GetDownloadState()) {
		t.Fatalf(
			`DB:GetState: Expected "%v", got "%v"`,
			states.GetStateStr(t2.GetDownloadState()),
			state,
		)
	}

	count, err := db.ResetFromCompletionState(states.TASK_STATE_COMPLETED)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatalf("DB:ResetFromCompletionState: Expected 2 rows affected, got %d", count)
	}

	err = db.Drop("*")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStateManipulation2(t *testing.T) {
	db := GetNewDatabase[*task.Task](DB_SQlite)

	err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	t1 := task.NewTask("hello")
	_, err = db.Insert(t1)
	if err != nil {
		t.Fatal(err)
	}

	state, err := db.AdvanceState(t1)
	if err != nil {
		t.Fatal(err)
	}
	t1.DownloadState = states.TASK_STATE_RUNNING

	if state != t1.DownloadState {
		t.Fatalf("Expected state %d, got %d", t1.DownloadState, state)
	}

	state, err = db.AdvanceState(t1)
	if err != nil {
		t.Fatal(err)
	}
	t1.DownloadState = states.TASK_STATE_COMPLETED

	if state != t1.DownloadState {
		t.Fatalf("Expected state %d, got %d", t1.DownloadState, state)
	}

	state, _ = db.AdvanceState(t1)
	if state != t1.DownloadState {
		t.Fatalf("Expected state %d, got %d", t1.DownloadState, state)
	}

	state, err = db.ResetState(t1)
	if err != nil {
		t.Fatal(err)
	}
	t1.DownloadState = states.TASK_STATE_QUEUED

	if state != t1.DownloadState {
		t.Fatalf("Expected state %d, got %d", t1.DownloadState, state)
	}

	state, _ = db.RegressState(t1)
	if state != t1.DownloadState {
		t.Fatalf("Expected state %d, got %d", t1.DownloadState, state)
	}

	err = db.Drop("*")
	if err != nil {
		t.Fatal(err)
	}
}
