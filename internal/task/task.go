package task

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
	SetDownloadState chan int
	// State progress percentage (from 0 to 100)
	Progress int8
	// Stores an eventual error occurred in the task lifecycle
	Err error
	// Aborts the task progression
	Stop chan struct{}
}

func NewTask() *Task {
	return &Task{}
}

func NewTaskFromServiceURL(aggregatorSlug string) *Task {
	return &Task{
		AggregatorSlug: aggregatorSlug,
	}
}

func NewTaskFromSlug(slug string) *Task {
	return &Task{
		AggregatorSlug: slug,
	}
}
