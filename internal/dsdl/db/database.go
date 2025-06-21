package db

import "github.com/relepega/doujinstyle-downloader/internal/task"

type DBType int

const (
	DB_Memory DBType = iota
	DB_SQlite
	max_db_enum
)

type DB[T task.Insertable] interface {
	Open() error
	Close() error
	Name() string

	// Returns the total number of stored tasks
	Count() (int, error)

	// Returns the total count of tasks in a specific completion state.
	//
	// Also returns an error if the specified completion state is invalid
	CountFromState(completionState int) (int, error)

	// Adds a task to the database
	Insert(nv T) error

	// Checks whether a task with an equal value is already present in the database
	Find(id string) (bool, error)

	// Checks whether a task with an equal value is already present in the database
	Get(slug string) (T, error)

	// Returns all the tasks in the database
	GetAll() ([]T, error)

	// Removes a task from the database
	//
	// Returns an error if trying to remove a task in a running state
	Remove(nv T) error

	// Removes multiple tasks with the same state from the database
	//
	// Returns the number of affected tasks and. If -1, then the state is out of range
	//
	// Also returns an error if something goes wrong while handling the database
	RemoveFromState(completionState int) (int, error)

	// Empties the database
	RemoveAll() error

	// Resets the state of EVERY task in the specified completion state
	//
	// Returns an error either if the completion state is invalid or if trying to reset tunning tasks
	ResetFromCompletionState(completionState int) error

	// Returns the state of a specific task. Returns an error if the task has not been found
	GetState(nv T) (string, error)

	// Sets the state of a specific task. Returns an error if the task has not been found
	SetState(nv T, newState int) error

	// Advances the completion state of a specific task
	//
	// Returns an error if the task has reached a completion state and the updated state value
	AdvanceState(nv T) (int, error)

	// Regresses the completion state of a specific task
	//
	// Returns an error if the task has reached a queued state and the updated state value
	RegressState(nv T) (int, error)

	// Resets the state of a specific task to a queued state
	//
	// Returns an error if trying to reset the state of a task in a running state and the updated state value
	ResetState(nv T) (int, error)

	// Drops specified table name
	Drop(table string) error
}

func GetNewDatabase[T task.Insertable](dbType DBType) DB[T] {
	switch dbType {
	case DB_Memory:
		return NewMemoryDB[T]()
	case DB_SQlite:
		return NewSQliteDB[T]()
	default:
		return NewMemoryDB[T]()
	}
}
