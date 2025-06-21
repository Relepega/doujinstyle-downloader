package db

import (
	"fmt"
	"strings"
	"sync"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

type MemoryDB[T task.Insertable] struct {
	DB[T]

	sync.Mutex

	name string

	tasks_db map[*T]int
}

func NewMemoryDB[T task.Insertable]() DB[T] {
	return &MemoryDB[T]{
		name: "In-MemoryDB",
		// tasks_db: make(map[make(*task.Task{})]int, 15) // seems a fair, arbitrary value
		tasks_db: make(map[*T]int, 15),
	}
}

func (db *MemoryDB[T]) Open() error {
	return nil
}

func (db *MemoryDB[T]) Close() error {
	return nil
}

func (db *MemoryDB[T]) Name() string {
	return db.name
}

func (db *MemoryDB[T]) Count() (int, error) {
	return len(db.tasks_db), nil
}

func (db *MemoryDB[T]) CountFromState(completionState int) (int, error) {
	db.Lock()
	defer db.Unlock()

	if completionState < 0 || completionState >= states.MaxCompletionState() {
		return -1, fmt.Errorf("CompletionState is not a value within constraints")
	}

	count := 0
	for _, v := range db.tasks_db {
		if v == completionState {
			count++
		}
	}

	return count, nil
}

func (db *MemoryDB[T]) Insert(nv T) error {
	db.Lock()
	defer db.Unlock()

	db.tasks_db[&nv] = states.TASK_STATE_QUEUED

	return nil
}

func (db *MemoryDB[T]) Find(id string) (bool, error) {
	db.Lock()
	defer db.Unlock()

	for k := range db.tasks_db {
		if (*k).GetID() == id {
			return true, nil
		}
	}

	return false, nil
}

func (db *MemoryDB[T]) Get(slug string) (T, error) {
	db.Lock()
	defer db.Unlock()

	var empty T

	for k := range db.tasks_db {
		if strings.Contains((*k).GetSlug(), slug) {
			return (*k), nil
		}
	}

	return empty, nil
}

func (db *MemoryDB[T]) GetAll() ([]T, error) {
	db.Lock()
	defer db.Unlock()

	// create a read-only copy of the database
	ro_db := make([]T, len(db.tasks_db))

	i := 0

	for k := range db.tasks_db {
		ro_db[i] = *k
		i++
	}

	return ro_db, nil
}

func (db *MemoryDB[T]) Remove(nv T) error {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if *k == nv {
			if v == states.TASK_STATE_RUNNING {
				return fmt.Errorf("Cannot remove a running task")
			}

			delete(db.tasks_db, k)
		}
	}

	return nil
}

func (db *MemoryDB[T]) RemoveFromState(completionState int) (int, error) {
	db.Lock()
	defer db.Unlock()

	if completionState < 0 || completionState >= states.MaxCompletionState() {
		return -1, nil
	}

	count := 0
	for k, v := range db.tasks_db {
		if v == completionState {
			delete(db.tasks_db, k)
		}
	}

	return count, nil
}

func (db *MemoryDB[T]) RemoveAll() error {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if v != states.TASK_STATE_RUNNING {
			delete(db.tasks_db, k)
		}
	}

	return nil
}

func (db *MemoryDB[T]) ResetFromCompletionState(completionState int) error {
	db.Lock()
	defer db.Unlock()

	if completionState < 0 || completionState >= states.MaxCompletionState() {
		return fmt.Errorf("Argument is not a valid state within constraints")
	}

	if completionState == states.TASK_STATE_RUNNING {
		return fmt.Errorf("Cannot cancel running tasks")
	}

	for k, v := range db.tasks_db {
		if v == completionState {
			delete(db.tasks_db, k)
		}
	}

	return nil
}

func (db *MemoryDB[T]) GetState(nv T) (string, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if *k == nv {
			return states.GetStateStr(v), nil
		}
	}

	return "", fmt.Errorf("Node not found")
}

func (db *MemoryDB[T]) SetState(nv T, newState int) error {
	db.Lock()
	defer db.Unlock()

	if newState < 0 || newState >= states.MaxCompletionState() {
		return fmt.Errorf("newState is out of bounds")
	}

	for k := range db.tasks_db {
		if *k == nv {
			db.tasks_db[k] = newState
			return nil
		}
	}

	return fmt.Errorf("Node not found")
}

func (db *MemoryDB[T]) AdvanceState(nv T) (int, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if *k == nv {
			if v >= states.MaxCompletionState() {
				return -1, fmt.Errorf("Cannot advance the status of this task anymore")
			}

			db.tasks_db[k]++
			return db.tasks_db[k], nil

		}
	}

	return -1, nil
}

func (db *MemoryDB[T]) RegressState(nv T) (int, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if *k == nv {
			if v <= 0 {
				return -1, fmt.Errorf("Cannot regress the status of this task anymore")
			}

			db.tasks_db[k]--
			return db.tasks_db[k], nil
		}
	}

	return -1, nil
}

func (db *MemoryDB[T]) ResetState(nv T) (int, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if *k == nv {
			if v == states.TASK_STATE_RUNNING {
				return -1, fmt.Errorf("Cannot reset an already running task")
			}

			db.tasks_db[k] = states.TASK_STATE_QUEUED
			return db.tasks_db[k], nil
		}
	}

	return -1, nil
}

func (db *MemoryDB[T]) Drop(table string) error {
	return db.RemoveAll()
}
