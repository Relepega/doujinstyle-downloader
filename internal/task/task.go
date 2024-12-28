package task

import (
	"fmt"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

type Task struct {
	// The actual Unique ID
	Id string
	// Aggregator formal name (e.g "doujinstyle")
	Aggregator string
	// Can be either the full url or the page id
	Slug string
	// Full URL calculated by combining name & slug
	AggregatorPageURL string
	// Filehost full url
	FilehostUrl string
	// Full name to be displayed on GUI
	DisplayName string
	// Downloaded filename
	Filename string
	// Mirror value of the one stored in the database
	DownloadState int
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
		DownloadState: dsdl.TASK_STATE_QUEUED,
		Stop:          make(chan struct{}),
	}

	return t
}

func (t *Task) SetProgress(p int8) {
	t.Progress = p
}

func (t *Task) ID() string {
	return t.Id
}
