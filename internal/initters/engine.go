package initters

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/aggregators"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/filehosts"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

func InitEngine(cfg *configManager.Config, ctx context.Context) *dsdl.DSDL {
	log.Println("starting playwright")
	pww, err := playwrightWrapper.UsePlaywright(
		"firefox",
		!cfg.Dev.PlaywrightDebug,
		0.0,
	)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("playwright started without errors")

	engine := dsdl.NewDSDLWithBrowser(ctx, pww.Browser)

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
			runningCount, err := tq.TrackerCountFromState(dsdl.TASK_STATE_RUNNING)
			if err != nil {
				continue
			}

			if tq.GetQueueLength() == 0 || runningCount == int(options.Download.ConcurrentJobs) {
				continue
			}

			taskVal, newState, err := tq.AdvanceNewTaskState()
			if err != nil {
				continue
			}

			taskData, ok := taskVal.(*task.Task)
			if !ok {
				panic("TaskRunner: Cannot convert node value into proper type\n")
			}
			taskData.DownloadState = newState

			appUtils.CreateAppTempDir(appUtils.GetAppTempDir())

			go taskRunner(tq, taskData, options.Download.Directory)
		}
	}
}

func taskRunner(tq *dsdl.TQProxy, taskData *task.Task, downloadPath string) {
	markCompleted := func() {
		newState, err := tq.AdvanceTaskState(taskData)
		if err != nil {
			panic(err)
		}
		taskData.DownloadState = newState
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
				taskData.Err = fmt.Errorf("Playwright: Cannot open new browser context")
				markCompleted()
				return
			}
			defer bwContext.Close()

			p, err := bwContext.NewPage()
			if err != nil {
				taskData.Err = fmt.Errorf("Playwright: Cannot open new browser context page")
				markCompleted()
				return
			}
			defer p.Close()

			aggregator := aggConstFn(taskData.AggregatorSlug, p)

			_, err = p.Goto(aggregator.Url())
			// check internet connection
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			// check if page is actually not deleted
			is404, err := aggregator.Is404()
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}
			if is404 {
				taskData.Err = fmt.Errorf(
					"Aggregator: The requested page has been taken down or is invalid",
				)
				markCompleted()
				return
			}

			// evaluate final filename
			fname, err := aggregator.EvaluateFileName()
			if err != nil {
				taskData.Err = fmt.Errorf("Aggregator: Couldn't evaluate the filename")
				markCompleted()
				return
			}

			fext, err := aggregator.EvaluateFileExt()
			if err != nil {
				taskData.Err = fmt.Errorf("Aggregator: Couldn't evaluate the file extension")
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
