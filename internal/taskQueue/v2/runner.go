package queue

import (
	"context"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
)

type RunQueueOpts struct {
	ctx            context.Context
	maxConcurrency int
	pwc            *playwrightWrapper.PwContainer
}

func RunQueue(
	tq *TQv2,
	opts interface{},
) {
	// parse options
	parsedOpts, ok := opts.(RunQueueOpts)
	if !ok {
		panic("RunQueue: Cannot parse function options")
	}

	// open empty page so that the context won't close
	emptyPage, _ := parsedOpts.pwc.BrowserContext.NewPage()
	defer emptyPage.Close()

	for {
		if tq.GetQueueLength() == 0 ||
			tq.TrackerCount(TASK_STATE_RUNNING) == parsedOpts.maxConcurrency {
			continue
		}

		taskVal, err := tq.AdvanceNewTaskState()
		if err != nil {
			continue
		}

		// TODO: complete with an actual Node struct value
		nodev, ok := taskVal.(NodeValue)
		if !ok {
			tq.RemoveNode(taskVal)
			continue
		}

		tq.AdvanceTaskState(nodev)

		go func(t *Tracker, v NodeValue) {
			time.Sleep(time.Second * 5)
			tq.AdvanceTaskState(v)
		}(tq.GetTracker(), nodev)
	}
}
