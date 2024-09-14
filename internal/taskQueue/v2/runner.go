package queue

import (
	"context"

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

	q := tq.GetQueue()
	t := tq.GetTracker()

	for {
		if q.Length() == 0 || t.Count(TASK_STATE_RUNNING) == parsedOpts.maxConcurrency {
			continue
		}

		taskVal, err := q.Dequeue()
		if err != nil {
			continue
		}

		nodev, ok := taskVal.(NodeValue)
		if !ok {
			continue
		}

		t.AdvanceState(nodev)

		go func(t *Tracker, v NodeValue) {
			return
		}(t, nodev)
	}
}
