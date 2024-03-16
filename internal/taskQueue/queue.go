package taskQueue

import (
	"context"
	"fmt"
	"sync"

	"github.com/relepega/doujinstyle-downloader/internal/downloader"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"

	"github.com/playwright-community/playwright-go"
)

type Queue struct {
	mu             sync.Mutex
	maxConcurrency int
	runningTasks   int
	tasks          []Task
	ctx            context.Context
	cancel         context.CancelFunc
}

type UIQueue struct {
	QueueLength int
	Tasks       []Task
}

func NewQueue(maxConcurrency int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		maxConcurrency: maxConcurrency,
		runningTasks:   0,
		tasks:          []Task{},
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (q *Queue) NewQueueFree() *UIQueue {
	return &UIQueue{
		QueueLength: q.GetQueueLength(),
		Tasks:       q.GetTasks(),
	}
}

func (q *Queue) AddTask(t *Task) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.tasks = append(q.tasks, *t)
}

func (q *Queue) RemoveTask(albumID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var newTaskList []Task
	for _, el := range q.tasks {
		if el.UrlSlug != albumID {
			newTaskList = append(newTaskList, el)
		}
	}

	q.tasks = newTaskList
}

func (q *Queue) GetTasks() []Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.tasks
}

func (q *Queue) GetQueueLength() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.tasks)
}

func (q *Queue) IsInList(t *Task) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, task := range q.tasks {
		if task.UrlSlug == t.UrlSlug {
			return true
		}
	}

	return false
}

func (q *Queue) ActivateFreeTask() (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

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
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, t := range q.tasks {
		if t.UrlSlug == albumID {
			q.tasks[i].Reset()
		}
	}
}

func (q *Queue) MarkTaskAsDone(t Task, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, rt := range q.tasks {
		if rt.UrlSlug == t.UrlSlug {
			q.tasks[i].MarkAsDone(err)
			q.runningTasks--
			return
		}
	}
}

func (q *Queue) ClearQueuedTasks() {
	q.mu.Lock()
	defer q.mu.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if t.Done || t.Active {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ClearSuccessfullyCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if !t.Done || t.Error != nil {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ClearFailedCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if !t.Done || t.Error == nil {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ClearAllCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	var filtered []Task

	for _, t := range q.tasks {
		if !t.Done {
			filtered = append(filtered, t)
		}
	}

	q.tasks = filtered
}

func (q *Queue) ResetFailedTasks() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, t := range q.tasks {
		if t.Done && (t.Error != nil) {
			q.tasks[i].Reset()
		}
	}
}

func (q *Queue) Run(pwc *playwrightWrapper.PwContainer) {
	// open empty page so that the context won't close
	emptyPage, _ := pwc.BrowserContext.NewPage()
	defer emptyPage.Close()

	for {
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

		go func(q *Queue, t *Task, bw *playwright.Browser) {
			select {
			case <-q.ctx.Done():
				// If the context is cancelled, break function
				return
			default:
				err := downloader.Download(t.UrlSlug, bw, &t.DownloadProgress, t.ServiceNumber)
				q.MarkTaskAsDone(*t, err)
			}
		}(q, task, &pwc.Browser)
	}
}

func (q *Queue) Shutdown(pwc *playwrightWrapper.PwContainer) {
	pwc.Close()

	// Cancel the context, which will stop all tasks
	q.cancel()
}
