package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/aggregators"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/filehosts"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue/task"
)

func initEngine(cfg *configManager.Config, ctx context.Context) *dsdl.DSDL {
	engine := dsdl.NewDSDL(ctx)

	engine.RegisterAggregator(&dsdl.Aggregator{
		Name:        "doujinstyle",
		Constructor: aggregators.NewDoujinstyle,
	})

	engine.RegisterAggregator(&dsdl.Aggregator{
		Name:        "sukidesuost",
		Constructor: aggregators.NewSukiDesuOst,
	})

	engine.RegisterFilehost(&dsdl.Filehost{
		Name:        "mediafire",
		Constructor: filehosts.NewMediafire,
	})

	engine.NewTQProxy(queueRunner)

	engine.Tq.RunQueue(cfg)

	return engine
}

func queueRunner(tq *dsdl.TQProxy, stop <-chan struct{}, opts interface{}) error {
	options, ok := opts.(configManager.Config)
	if !ok {
		log.Fatalln("Options are of wrong type")
	}

	for {
		select {
		case <-stop:
			return nil

		default:
			tcount, err := tq.TrackerCountFromState(dsdl.TASK_STATE_RUNNING)
			if err != nil {
				continue
			}

			if tq.GetQueueLength() == 0 || tcount == int(options.Download.ConcurrentJobs) {
				time.Sleep(time.Millisecond)
				continue
			}

			taskVal, err := tq.AdvanceNewTaskState()
			if err != nil {
				continue
			}

			taskData, ok := taskVal.(*task.Task)
			if !ok {
				panic("TaskRunner: Cannot convert node value into proper type\n")
			}
			taskData.SetDownloadState <- dsdl.TASK_STATE_RUNNING

			go taskRunner(tq, taskData)
		}
	}
}

func taskRunner(tq *dsdl.TQProxy, taskData *task.Task) {
	markCompleted := func() {
		err := tq.AdvanceTaskState(taskData)
		if err != nil {
			panic(err)
		}
		taskData.SetDownloadState <- dsdl.TASK_STATE_COMPLETED
	}

	engine := tq.Context()

	running := false

	for {
		select {
		case <-taskData.Stop:
			taskData.Err = fmt.Errorf("task aborted by the user")
			markCompleted()
			return

		default:
			if running {
				continue
			}

			// mark running, so that we don't end with a memory leak :)
			running = true

			// process the task
			aggConstFn, err := engine.EvaluateAggregator(taskData.AggregatorName)
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			bwContext, err := engine.GetBrowserInstance().NewContext()
			if err != nil {
				taskData.Err = fmt.Errorf("taskRunner: Cannot open new browser context")
				markCompleted()
				return
			}

			p, err := bwContext.NewPage()
			if err != nil {
				taskData.Err = fmt.Errorf("taskRunner: Cannot open new browser context page")
				markCompleted()
				return
			}

			aggregator := aggConstFn(taskData.AggregatorSlug)

			// task done :)
			markCompleted()
		}
	}
}
