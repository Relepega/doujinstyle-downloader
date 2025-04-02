package database

// Emun of completion states. Used to track a task's state.
const (
	TASK_STATE_QUEUED int = iota
	TASK_STATE_RUNNING
	TASK_STATE_COMPLETED
	max_completion_state
)

func MaxCompletionState() int {
	return max_completion_state
}

// String rappresentation & meaning for every completion state.
//
// The can be accessed through t.GetState()
const (
	TASK_STATE_QUEUED_STR    = "Queued"
	TASK_STATE_RUNNING_STR   = "Running"
	TASK_STATE_COMPLETED_STR = "Completed"
)

var statuses = map[int]string{
	TASK_STATE_QUEUED:    TASK_STATE_QUEUED_STR,
	TASK_STATE_RUNNING:   TASK_STATE_RUNNING_STR,
	TASK_STATE_COMPLETED: TASK_STATE_COMPLETED_STR,
}

func GetStateStr(state int) string {
	return statuses[state]
}

type DBType int

const (
	DB_SQlite DBType = iota
	DB_Memory
	max_db_enum
)

type DB interface {
	// Returns the total number of stored tasks
	Count() (int, error)

	// Returns the total count of tasks in a specific completion state.
	//
	// Also returns an error if the specified completion state is invalid
	CountFromState(completionState int) (int, error)

	// Adds a task to the database
	Add(nv any) error

	// Checks whether a task with an equal value is already present in the database
	Get(nv any) (bool, error)

	// Returns all the tasks in the database
	GetAll() (map[any]int, error)

	// Removes a task from the database
	//
	// Returns an error if trying to remove a task in a running state
	Remove(nv any) error

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
	GetState(nv any) (string, error)

	// Sets the state of a specific task. Returns an error if the task has not been found
	SetState(nv any, newState int) error

	// Advances the completion state of a specific task
	//
	// Returns an error if the task has reached a completion state and the updated state value
	AdvanceState(nv any) (int, error)

	// Regresses the completion state of a specific task
	//
	// Returns an error if the task has reached a queued state and the updated state value
	RegressState(nv any) (int, error)

	// Resets the state of a specific task to a queued state
	//
	// Returns an error if trying to reset the state of a task in a running state and the updated state value
	ResetState(nv any) (int, error)
}

func GetNewDatabase(dbType DBType) DB {
	switch dbType {
	case 0:
		return NewMemoryDB()
	default:
		return NewMemoryDB()
	}
}
