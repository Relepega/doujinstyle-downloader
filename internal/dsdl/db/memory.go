package database

import (
	"fmt"
	"sync"
)

type MemoryDB struct {
	DB

	sync.Mutex

	tasks_db map[any]int
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		tasks_db: make(map[any]int, 15), // seems a fair, arbitrary value
	}
}

func (db *MemoryDB) Count() (int, error) {
	return len(db.tasks_db), nil
}

func (db *MemoryDB) CountFromState(completionState int) (int, error) {
	db.Lock()
	defer db.Unlock()

	if completionState < 0 || completionState >= max_completion_state {
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

func (db *MemoryDB) Add(nv any) error {
	db.Lock()
	defer db.Unlock()

	db.tasks_db[nv] = TASK_STATE_QUEUED

	return nil
}

func (db *MemoryDB) Get(nv any) (bool, error) {
	db.Lock()
	defer db.Unlock()

	for k := range db.tasks_db {
		if k == nv {
			return true, nil
		}
	}
	return false, nil
}

func (db *MemoryDB) GetAll() (map[any]int, error) {
	db.Lock()
	defer db.Unlock()

	return db.tasks_db, nil
}

func (db *MemoryDB) Remove(nv any) error {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if k == nv {
			if v == TASK_STATE_RUNNING {
				return fmt.Errorf("Cannot remove a running task")
			}

			delete(db.tasks_db, k)
		}
	}

	return nil
}

func (db *MemoryDB) RemoveFromState(completionState int) (int, error) {
	db.Lock()
	defer db.Unlock()

	if completionState < 0 || completionState >= max_completion_state {
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

func (db *MemoryDB) RemoveAll() error {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if v != TASK_STATE_RUNNING {
			delete(db.tasks_db, k)
		}
	}

	return nil
}

func (db *MemoryDB) ResetFromCompletionState(completionState int) error {
	db.Lock()
	defer db.Unlock()

	if completionState < 0 || completionState >= max_completion_state {
		return fmt.Errorf("Argument is not a valid state within constraints")
	}

	if completionState == TASK_STATE_RUNNING {
		return fmt.Errorf("Cannot cancel running tasks")
	}

	for k, v := range db.tasks_db {
		if v == completionState {
			delete(db.tasks_db, k)
		}
	}

	return nil
}

func (db *MemoryDB) GetState(nv any) (string, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if k == nv {
			return statuses[v], nil
		}
	}

	return "", fmt.Errorf("Node not found")
}

func (db *MemoryDB) SetState(nv any, newState int) error {
	db.Lock()
	defer db.Unlock()

	if newState < 0 || newState >= max_completion_state {
		return fmt.Errorf("newState is out of bounds")
	}

	for k := range db.tasks_db {
		if k == nv {
			db.tasks_db[k] = newState
			return nil
		}
	}

	return fmt.Errorf("Node not found")
}

func (db *MemoryDB) AdvanceState(nv any) (int, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if k == nv {
			if v >= max_completion_state {
				return -1, fmt.Errorf("Cannot advance the status of this task anymore")
			}

			db.tasks_db[k]++
			return db.tasks_db[k], nil

		}
	}

	return -1, nil
}

func (db *MemoryDB) RegressState(nv any) (int, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if k == nv {
			if v <= 0 {
				return -1, fmt.Errorf("Cannot regress the status of this task anymore")
			}

			db.tasks_db[k]--
			return db.tasks_db[k], nil
		}
	}

	return -1, nil
}

func (db *MemoryDB) ResetState(nv any) (int, error) {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.tasks_db {
		if k == nv {
			if v == TASK_STATE_RUNNING {
				return -1, fmt.Errorf("Cannot reset an already running task")
			}

			db.tasks_db[k] = TASK_STATE_QUEUED
			return db.tasks_db[k], nil
		}
	}

	return -1, nil
}
