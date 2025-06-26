package states

// Emun of completion states. Used to track a task's state.
const (
	TASK_STATE_QUEUED int = iota
	TASK_STATE_RUNNING
	TASK_STATE_COMPLETED
	max_completion_state
)

func MaxCompletionState() int {
	return max_completion_state - 1
}

// String rappresentation & meaning for every completion state.
//
// The can be accessed through t.GetState()
const (
	TASK_STATE_QUEUED_STR    = "Queued"
	TASK_STATE_RUNNING_STR   = "Running"
	TASK_STATE_COMPLETED_STR = "Completed"
)

var statesMap = map[int]string{
	TASK_STATE_QUEUED:    TASK_STATE_QUEUED_STR,
	TASK_STATE_RUNNING:   TASK_STATE_RUNNING_STR,
	TASK_STATE_COMPLETED: TASK_STATE_COMPLETED_STR,
}

func GetStateStr(state int) string {
	return statesMap[state]
}
