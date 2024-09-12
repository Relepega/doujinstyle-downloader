package queue

import (
	"context"
	"log"

	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	tq_eventbroker "github.com/relepega/doujinstyle-downloader/internal/taskQueue/tq_event_broker"
)

func RunQueue[T any](q *Queue[T], ctx context.Context, pwc *playwrightWrapper.PwContainer) {
	// open empty page so that the context won't close
	emptyPage, _ := pwc.BrowserContext.NewPage()
	defer emptyPage.Close()

	queue_pub := pubsub.NewGlobalPublisher("queue")
	subscriber := queue_pub.Subscribe()

	for {
		select {
		case <-ctx.Done():
			// quit all the ongoing tasks and then return
			log.Println("Graceful queue shutdown complete.")
			return

		case evt := <-subscriber:
			switch evt.EvtType {
			case "update-task-progress":
				evt_data := evt.Data.(*tq_eventbroker.UpdateTaskProgress)

				t, err := q.GetTask(evt_data.Id)
				if err != nil {
					continue
				}

				t.DownloadProgress = evt_data.Progress
				q.publishUIUpdate("update-task-content", t)

			default:
				continue
			}
		default:
			// run task scheduler
			if (q.runningTasks == q.maxConcurrency) || (len(q.tasks) == 0) {
				continue
			}

			task, err := q.ActivateFreeTask()
			if err != nil {
				continue
			}

			if task == nil {
				continue
			}

			go func(q *Queue, t *Task, pwc *playwrightWrapper.PwContainer) {
				err := t.Run(q, pwc)
				q.MarkTaskAsDone(*t, err)
			}(q, task, pwc)
		}
	}
}
