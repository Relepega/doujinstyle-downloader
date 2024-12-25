package task

import (
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

type Task struct {
	// Aggregator formal name (e.g "doujinstyle")
	AggregatorName string
	// Can be either the full url or the page id
	AggregatorSlug string
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
	return &Task{
		AggregatorSlug: slug,
		DisplayName:    slug,
		DownloadState:  dsdl.TASK_STATE_QUEUED,
		Stop:           make(chan struct{}),
	}
}

func (t *Task) SetProgress(p int8) {
	t.Progress = p
}
