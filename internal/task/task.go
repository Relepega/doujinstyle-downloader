package task

import (
	"fmt"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
)

type Task struct {
	// The actual Unique ID
	Id string `db:"ID"`
	// Aggregator formal name (e.g "doujinstyle")
	Aggregator string `db:"Aggregator"`
	// Can be either the full url or the page id
	Slug string `db:"Slug"`
	// Full URL calculated by combining name & slug
	AggregatorPageURL string `db:"AggregatorPageURL"`
	// Filehost full url
	FilehostUrl string `db:"FilehostUrl"`
	// Full name to be displayed on GUI
	DisplayName string `db:"DisplayName"`
	// Downloaded filename
	Filename string `db:"Filename"`
	// Mirror value of the one stored in the database
	DownloadState int `db:"DownloadState"`
	// State progress percentage (from -1 (not yet downloading) to 100)
	Progress int8
	// Stores an eventual error occurred in the task lifecycle
	Err error
	// Aborts the task progression
	Stop chan struct{}
}

func NewTask(slug string) *Task {
	t := &Task{
		Id: fmt.Sprintf(
			"%d-%s",
			time.Now().UnixMilli(),
			appUtils.GenerateRandomFilename(),
		),
		Slug:          slug,
		DisplayName:   slug,
		DownloadState: states.TASK_STATE_QUEUED,
		Progress:      -1,
		Stop:          make(chan struct{}),
	}

	return t
}

func (t *Task) ID() string { return t.Id }

func (t *Task) SetState(state int) { t.DownloadState = state }

func (t *Task) SetProgress(p int8) { t.Progress = p }

func (t *Task) SetErrMsg(m string) { t.Err = fmt.Errorf("%s", m) }

func (t *Task) SetErr(err error) { t.Err = err }

func (t *Task) Abort() {
	t.Stop <- struct{}{}
}
