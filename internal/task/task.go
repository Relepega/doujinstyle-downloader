package task

import (
	"fmt"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
)

type Insertable interface {
	comparable

	GetID() string
	GetAggregator() string
	GetSlug() string
	GetAggregatorPageURL() string
	GetFilehostUrl() string
	GetDisplayName() string
	GetFilename() string
	GetDownloadState() int
	GetProgress() int
	GetErrMsg() string
	compare(c any) int
}

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

type DBRecord struct {
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

func (t *Task) GetID() string { return t.Id }

func (t *Task) GetAggregator() string { return t.Aggregator }

func (t *Task) GetSlug() string { return t.Slug }

func (t *Task) GetAggregatorPageURL() string { return t.AggregatorPageURL }

func (t *Task) GetFilehostUrl() string { return t.FilehostUrl }

func (t *Task) GetDisplayName() string { return t.DisplayName }

func (t *Task) GetFilename() string { return t.Filename }

func (t *Task) GetDownloadState() int { return t.DownloadState }

func (t *Task) GetProgress() int { return int(t.Progress) }

func (t *Task) GetErrMsg() string {
	if t.Err != nil {
		return t.Err.Error()
	}

	return ""
}

func (t *Task) compare(c any) int {
	cv, ok := c.(*Task) //  getting  the instance of T via type assertion.
	if !ok {
		return -1
	}

	if cv != t {
		return 0
	}

	return 1
}
