package taskQueue

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/relepega/doujinstyle-downloader/internal/downloader"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"

	"github.com/playwright-community/playwright-go"
)

type Queue struct {
	mu             sync.Mutex
	Interrupt      chan os.Signal
	maxConcurrency int
	runningTasks   int
	tasks          []Task
}

type UIQueue struct {
	QueueLength int
	Tasks       []Task
}

func NewQueue(maxConcurrency int) *Queue {
	return &Queue{
		Interrupt:      make(chan os.Signal, 1),
		maxConcurrency: maxConcurrency,
		runningTasks:   0,
		tasks:          []Task{},
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
		if el.AlbumID != albumID {
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
		if task.AlbumID == t.AlbumID {
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
		if t.AlbumID == albumID {
			q.tasks[i].Reset()
		}
	}
}

func (q *Queue) MarkTaskAsDone(t Task, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, rt := range q.tasks {
		if rt.AlbumID == t.AlbumID {
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

func (q *Queue) Run() {
	pw, bw, ctx, err := playwrightWrapper.UsePlaywright(
		playwrightWrapper.WithBrowserType(),
		playwrightWrapper.WithHeadless(),
		playwrightWrapper.WithTimeout(),
	)
	if err != nil {
		log.Fatalf("Queue.Run() error: Cannot open playwright browser: %v", err)
	}

	// open empty page so that the context won't close
	emptyPage, _ := ctx.NewPage()

	for {
		select {
		case <-q.Interrupt:
			emptyPage.Close()
			playwrightWrapper.ClosePlaywright(pw, bw)
			return
		default:
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
				err := downloader.Download(t.AlbumID, bw)
				q.MarkTaskAsDone(*t, err)
			}(q, task, &bw)

		}
	}
}
