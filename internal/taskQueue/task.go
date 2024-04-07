package taskQueue

import (
	"fmt"

	"github.com/relepega/doujinstyle-downloader-reloaded/internal/playwrightWrapper"
)

type Task struct {
	AlbumID string
	Service string
	Active  bool
	Done    bool
	Error   error

	DownloadProgress int8
}

func NewTask(AlbumID string, ServiceNumber int) *Task {
	service := "" // TODO

	return &Task{
		AlbumID: AlbumID,
		Service: service,
		Active:  false,
		Done:    false,
		Error:   nil,

		DownloadProgress: -1,
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
	if t.Active {
		return
	}

	t.Active = false
	t.Done = false
	t.Error = nil
}

func (t *Task) SetDownloadProgress(p int8) {
	// TODO
}

func (t *Task) Run(pwc *playwrightWrapper.PwContainer) error {
	// TODO

	fmt.Println("t.Run(): Running task...")

	return nil
}
