package taskQueue

type Task struct {
	Active  bool
	Done    bool
	Error   error
	AlbumID string
}

func NewTask(s string) *Task {
	return &Task{
		Active:  false,
		Done:    false,
		Error:   nil,
		AlbumID: s,
	}
}

func (t *Task) Activate() {
	t.Active = true
}

func (t *Task) Deactivate() {
	t.Active = false
}

func (t *Task) MarkAsDone(e error) {
	t.Active = false
	t.Done = true
	t.Error = e
}

func (t *Task) Reset() {
	t.Active = false
	t.Done = false
	t.Error = nil
}
