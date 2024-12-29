package initters

import (
	"context"
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"

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
		&cfg.Download.Tempdir,
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
		Name:                "Mediafire",
		AllowedUrlWildcards: []string{"www.mediafire.com"},
		Constructor:         filehosts.NewMediafire,
	})

	engine.RegisterFilehost(&dsdl.Filehost{
		Name:                "Mega",
		AllowedUrlWildcards: []string{"mega.nz"},
		Constructor:         filehosts.NewMega,
	})

	engine.NewTQProxy(queueRunner)

	engine.GetTQProxy().SetComparatorFunc(func(item, target interface{}) bool {
		t := target.(*task.Task)
		dbTask := item.(*task.Task)

		if dbTask.ID() == t.ID() {
			return false
		}

		return dbTask.Slug == t.Slug ||
			dbTask.AggregatorPageURL == t.AggregatorPageURL
	})

	engine.Tq.RunQueue(cfg)

	return engine
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

			go taskRunner(tq, taskData, options.Download.Directory, options.Download.Tempdir)
		}
	}
}

func taskRunner(tq *dsdl.TQProxy, taskData *task.Task, downloadDir string, tempDir string) {
	var bwContext playwright.BrowserContext

	markCompleted := func() {
		newState, err := tq.AdvanceTaskState(taskData)
		if err != nil {
			panic(err)
		}
		taskData.DownloadState = newState
		bwContext.Close()
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
			aggConstFn, err := engine.EvaluateAggregator(taskData.Aggregator)
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			bwContext, err = engine.GetBrowserInstance().NewContext()
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

			aggregator := aggConstFn(taskData.Slug, p)

			taskData.AggregatorPageURL = aggregator.Url()

			_, err = p.Goto(aggregator.Url())
			// check internet connection
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			taskData.Slug = aggregator.Slug()

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

			// evaluate displayName filename
			fname, err := aggregator.EvaluateFileName()
			if fname != "" {
				taskData.DisplayName = fname
			}

			// get download page
			dlPage, err := aggregator.EvaluateDownloadPage()
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}
			defer dlPage.Close()

			// parse a filehost downloader
			filehostConstructor, err := engine.EvaluateFilehost(dlPage.URL())
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}
			filehost := filehostConstructor(dlPage)

			taskData.FilehostUrl = filehost.Page().URL()

			// evaluate final filename
			if fname == "" {
				fname, err = filehost.EvaluateFileName()
				if err != nil {
					taskData.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the filename")
					markCompleted()
					return
				}

				// setting the filename only if it is stil not set
				taskData.DisplayName = fname
			}

			fext, err := aggregator.EvaluateFileExt()
			if err != nil {
				fext, err = filehost.EvaluateFileExt()
				if err != nil {

					taskData.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the file extension")
					markCompleted()
					return
				}
			}

			// re-check if task is already done by other means
			found, _ := tq.Find(taskData)
			if found {
				taskData.Err = fmt.Errorf("This task is already present in the database")
				markCompleted()
				return
			}

			// check if out dirs exist
			if !appUtils.DirectoryExists(downloadDir) {
				err := appUtils.MkdirAll(downloadDir)
				if err != nil {
					log.Fatalln("taskRunner (dir_check):", err)
				}
			}

			if !appUtils.DirectoryExists(tempDir) {
				err := appUtils.MkdirAll(tempDir)
				if err != nil {
					log.Fatalln("taskRunner (dir_check):", err)
				}
			}

			// download the file into temp
			fullFilename := fmt.Sprintf("%s.%s", fname, fext)

			err = filehost.Download(tempDir, downloadDir, fullFilename, &taskData.Progress)
			if err != nil {
				taskData.Err = err
				markCompleted()
				return
			}

			// task done :)
			markCompleted()
		}
	}
}
