package task

import "github.com/relepega/doujinstyle-downloader/internal/dsdl"

type Task struct {
	// Aggregator formal name (e.g "doujinstyle")
	AggregatorName string
	// Can be either the full url or the page id
	AggregatorSlug string
	// Filehost full url
	FilehostUrl string
	// Downloaded filename
	Filename string
	// Sets the download state (e.g. "Downloading", "queued", "moving", ...) to the database
	SetDownloadState chan *dsdl.UpdateTaskDownloadState
	//
	DownloadState int
	// State progress percentage (from 0 to 100)
	Progress int8
	// Stores an eventual error occurred in the task lifecycle
	Err error
	// Aborts the task progression
	Stop chan struct{}
}

func NewTask(setDownloadStateChan chan *dsdl.UpdateTaskDownloadState) *Task {
	return &Task{
		SetDownloadState: setDownloadStateChan,
		DownloadState:    dsdl.TASK_STATE_QUEUED,
	}
}

func NewTaskFromServiceURL(
	setDownloadStateChan chan *dsdl.UpdateTaskDownloadState,
	aggregatorSlug string,
) *Task {
	return &Task{
		AggregatorSlug:   aggregatorSlug,
		SetDownloadState: setDownloadStateChan,
		DownloadState:    dsdl.TASK_STATE_QUEUED,
	}
}

func NewTaskFromSlug(setDownloadStateChan chan *dsdl.UpdateTaskDownloadState, slug string) *Task {
	return &Task{
		AggregatorSlug:   slug,
		SetDownloadState: setDownloadStateChan,
		DownloadState:    dsdl.TASK_STATE_QUEUED,
	}
}

func (t *Task) SetProgress(p int8) {
	t.Progress = p
}
