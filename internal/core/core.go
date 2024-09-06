package core

import (
	"sync"

	"github.com/relepega/doujinstyle-downloader/internal/core/taskQueue"
)

type DSDLCore struct {
	sync.RWMutex

	queue *taskQueue.Queue
}

func NewDSDLCore(maxConcurrency int8) *DSDLCore {
	return &DSDLCore{
		queue: taskQueue.NewQueue(maxConcurrency),
	}
}
