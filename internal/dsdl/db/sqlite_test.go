package db

import (
	"strings"
	"testing"

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

	err = db.Insert(task.NewTask("sqlite"))
	if err != nil {
		t.Fatal(err)
	}

	tsk, err := db.Get("ite")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains((*tsk).GetSlug(), "ite") {
		t.Fatal("Wrong task found")
	}

	err = db.Drop("*")
	if err != nil {
		t.Fatal(err)
	}
}
