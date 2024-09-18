package taskQueue

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	tq_eventbroker "github.com/relepega/doujinstyle-downloader/internal/taskQueue/tq_event_broker"
)

type Queue struct {
	tasks          []Task
	lock           sync.RWMutex
	runningTasks   int8
	maxConcurrency int8

	pub *pubsub.Publisher
}

type UIQueueData struct {
	QueueLength int
	Tasks       []Task
}

func NewQueue(MaxConcurrency int8, publisher *pubsub.Publisher) *Queue {
	return &Queue{
		tasks:          make([]Task, 0),
		runningTasks:   0,
		maxConcurrency: MaxConcurrency,

		pub: publisher,
	}
}

func (q *Queue) GetUIData() *UIQueueData {
	return &UIQueueData{
		QueueLength: q.GetQueueLength(),
		Tasks:       q.GetTasks(),
	}
}

func (q *Queue) GetQueueLength() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.tasks)
}

func (q *Queue) isTaskInList(task *Task) bool {
	for _, t := range q.tasks {
		if (t.AlbumID == task.AlbumID || t.DisplayName == task.DisplayName) && !t.IsChecking {
			return true
		}
	}
	return false
}

func (q *Queue) AddTask(task *Task) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.isTaskInList(task) {
		return fmt.Errorf("task already in list")
	}

	q.tasks = append(q.tasks, *task)

	return nil
}

func (q *Queue) RemoveTask(task *Task) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if !q.isTaskInList(task) {
		return fmt.Errorf("task not in list")
	}

	for i, t := range q.tasks {
		// remove only if inactive
		if t.AlbumID == task.AlbumID && !t.Active {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			break
		}
	}

	return nil
}

func (q *Queue) RemoveTaskFromAlbumID(AlbumID string) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	for i, t := range q.tasks {
		if t.AlbumID == AlbumID {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("task not in list")
}

func (q *Queue) GetTasks() []Task {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.tasks
}

func (q *Queue) GetTask(albumID string) (*Task, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	for _, t := range q.tasks {
		if t.AlbumID == albumID {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("task not in list")
}

func (q *Queue) GetLength() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.tasks)
}

func (q *Queue) ActivateFreeTask() (*Task, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	for i, t := range q.tasks {
		if !t.Active && !t.Done {
			q.tasks[i].Activate()
			q.runningTasks++

			return &q.tasks[i], nil
		}
	}

	return nil, fmt.Errorf("No free tasks")
}

func (q *Queue) ResetTask(albumID string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	for i, t := range q.tasks {
		if t.AlbumID == albumID {
			q.tasks[i].Reset()
		}
	}
}

func (q *Queue) MarkTaskAsDone(t Task, err error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	for i, task := range q.tasks {
		if task.AlbumID == t.AlbumID {
			q.tasks[i].MarkAsDone(err)
			q.runningTasks--

			q.publishUIUpdate("mark-task-as-done", &q.tasks[i])

			return
		}
	}
}

func (q *Queue) ClearQueuedTasks() {
	q.lock.Lock()
	defer q.lock.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if t.Done || t.Active {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ClearSuccessfullyCompleted() {
	q.lock.Lock()
	defer q.lock.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if !t.Done || t.Error != nil {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ClearFailedCompleted() {
	q.lock.Lock()
	defer q.lock.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if !t.Done || t.Error == nil {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ClearAllCompleted() {
	q.lock.Lock()
	defer q.lock.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if !t.Done {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ResetFailed() {
	q.lock.Lock()
	defer q.lock.Unlock()

	for i, t := range q.tasks {
		if t.Done && (t.Error != nil) {
			q.tasks[i].Reset()
		}
	}
}

func (q *Queue) publishUIUpdate(evt string, data interface{}) {
	q.pub.Publish(&pubsub.PublishEvent{
		EvtType: evt,
		Data:    data,
	})
}

func (q *Queue) Run(ctx context.Context, pwc *playwrightWrapper.PwContainer) {
	// open empty page so that the context won't close
	_, err := pwc.BrowserContext.NewPage()
	if err != nil {
		log.Panicf("Queue runner couldn't open a new browser page: %v\n", err)
	}
	// defer emptyPage.Close()

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
