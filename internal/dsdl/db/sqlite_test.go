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

	err = db.Insert(task.NewTask("hello"))
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

	err = db.Insert(task.NewTask("hello"))
	if err != nil {
		t.Fatal(err)
	}

	err = db.Insert(task.NewTask("world"))
	if err != nil {
		t.Fatal(err)
	}

	t1 := task.NewTask("sqlite")
	err = db.Insert(t1)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Insert(task.NewTask("sqlite"))
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

	err = db.Insert(task.NewTask("hello"))
	if err != nil {
		t.Fatal(err)
	}

	t1 := task.NewTask("world")
	err = db.Insert(t1)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Insert(task.NewTask("sqlite"))
	if err != nil {
		t.Fatal(err)
	}

	t2 := task.NewTask("apples")
	t2.DownloadState = states.TASK_STATE_RUNNING

	err = db.Insert(t2)
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
}
