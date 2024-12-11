package initters

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/aggregators"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/filehosts"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue/task"
)

func InitEngine(cfg *configManager.Config, ctx context.Context) *dsdl.DSDL {
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

func generateRandomFilename() (string, error) {
	// Generate a random string of 16 characters
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Convert the bytes to a hex string
	filename := fmt.Sprintf("%x", b)

	return filename, nil
}

func queueRunner(tq *dsdl.TQProxy, stop <-chan struct{}, opts interface{}) error {
	options, ok := opts.(*configManager.Config)
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

			appUtils.CreateAppTempDir(appUtils.GetAppTempDir())

			go taskRunner(tq, taskData, options.Download.Directory)
		}
	}
}

func taskRunner(tq *dsdl.TQProxy, taskData *task.Task, downloadPath string) {
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

			aggregator := aggConstFn(taskData.AggregatorSlug, p)

			// check if page is actually not deleted
			isValidPage, err := aggregator.Is404()
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}
			if !isValidPage {
				taskData.Err = fmt.Errorf(
					"taskRunner: The requested page has been taken down or is invalid",
				)
				markCompleted()
				return
			}

			// evaluate final filename
			fname, err := aggregator.EvaluateFileName()
			if err != nil {
				taskData.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the filename")
				markCompleted()
				return
			}

			fext, err := aggregator.EvaluateFileExt()
			if err != nil {
				taskData.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the file extension")
				markCompleted()
				return
			}

			finalfp := filepath.Join(downloadPath, fmt.Sprintf("%s.%s", fname, fext))

			// get download page
			dlPage, err := aggregator.EvaluateDownloadPage()
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			// parse a filehost downloader
			filehost, err := engine.EvaluateFilehost(dlPage.URL())
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			// TODO: edge cases on which the filename couldn't be evaluated from the aggregatorPage are not handled

			// make a temp filename to download into
			tempfn, err := generateRandomFilename()
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			tempfp := filepath.Join(appUtils.GetAppTempDir(), tempfn)

			// download the file into temp
			err = filehost.Download(tempfp, &taskData.Progress)
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			// move the temp file into final file
			os.Rename(tempfp, finalfp)

			// task done :)
			markCompleted()
		}
	}
}
