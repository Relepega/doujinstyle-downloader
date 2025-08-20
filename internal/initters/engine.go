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
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

func InitEngine(cfg *configManager.Config, ctx context.Context) *dsdl.DSDL {
	log.Println("starting playwright")
	pww, err := playwrightWrapper.UsePlaywright(
		&playwrightWrapper.PlaywrightOpts{
			BrowserType:   "firefox",
			Headless:      !cfg.Dev.PlaywrightDebug,
			Timeout:       0.0,
			DownloadsPath: cfg.Download.Tempdir,
		},
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

	engine.RegisterFilehost(&dsdl.Filehost{
		Name:                "Google Drive",
		AllowedUrlWildcards: []string{"drive.google.com"},
		Constructor:         filehosts.NewGDrive,
	})

	engine.RegisterFilehost(&dsdl.Filehost{
		Name:                "Jottacloud",
		AllowedUrlWildcards: []string{"jottacloud.com"},
		Constructor:         filehosts.NewJottacloud,
	})

	engine.NewTQProxy(queueRunner)
	// fmt.Println(engine.GetTQProxy().GetDatabase().Name())

	engine.GetTQProxy().SetComparatorFunc(func(item, target any) bool {
		t := target.(*task.Task)
		dbTask := item.(*task.Task)

		if dbTask.Id == t.Id {
			return false
		}

		return dbTask.Slug == t.Slug ||
			dbTask.AggregatorPageURL == t.AggregatorPageURL
	})

	engine.Tq.RunQueue(cfg)

	return engine
}

func queueRunner(tq *dsdl.TQProxy, stop <-chan struct{}, opts any) error {
	defer tq.GetDatabase().Close()

	options, ok := opts.(*configManager.Config)
	if !ok {
		log.Fatalln("Options are of wrong type")
	}

	maxJobs := int(options.Download.ConcurrentJobs)

	for {
		select {
		case <-stop:
			return nil

		default:
			if tq.GetQueueLength() == 0 || tq.GetActiveJobsCount() == maxJobs {
				continue
			}

			t, err := tq.Dequeue()
			if err != nil {
				continue
			}

			newState, err := tq.AdvanceTaskState(t)
			if err != nil {
				continue
			}

			t.DownloadState = newState

			go taskRunner(tq, t, options.Download.Directory, options.Download.Tempdir)
		}
	}
}

func taskRunner(
	tq *dsdl.TQProxy,
	activeTask *task.Task,
	downloadDir string,
	tempDir string,
) {
	var bwContext playwright.BrowserContext
	var publisher *pubsub.Publisher

	publisher, err := pubsub.GetGlobalPublisher("task-updater")
	if err != nil {
		publisher = pubsub.NewGlobalPublisher("task-updater")
	}

	markCompleted := func() {
		bwContext.Close()

		_, err := tq.AdvanceTaskState(activeTask)
		if err != nil {
			panic(err)
		}

		publisher.Publish(&pubsub.PublishEvent{
			EvtType: "mark-task-as-done",
			Data:    activeTask,
		})
	}

	engine := tq.Context()

	publisher.Publish(&pubsub.PublishEvent{
		EvtType: "activate-task",
		Data:    activeTask,
	})

	running := false

	for {
		select {
		case <-activeTask.Stop:
			activeTask.Err = fmt.Errorf("task aborted by the user")
			markCompleted()

			return

		default:
			if running {
				continue
			}

			// mark running, so that we don't end with a memory leak :)
			running = true

			// process the task
			aggConstFn, err := engine.EvaluateAggregator(activeTask.Aggregator)
			if err != nil {
				activeTask.Err = err
				markCompleted()
				return
			}

			bwContext, err = engine.GetBrowserInstance().NewContext()
			if err != nil {
				activeTask.Err = fmt.Errorf("Playwright: Cannot open new browser context")
				markCompleted()
				return
			}
			defer bwContext.Close()

			p, err := bwContext.NewPage()
			if err != nil {
				activeTask.Err = fmt.Errorf("Playwright: Cannot open new browser context page")
				markCompleted()
				return
			}
			defer p.Close()

			aggregator := aggConstFn(activeTask.Slug, p)

			activeTask.AggregatorPageURL = aggregator.Url()

			_, err = p.Goto(aggregator.Url())
			// check internet connection
			if err != nil {
				activeTask.Err = err
				markCompleted()
				return
			}

			activeTask.Slug = aggregator.Slug()

			// check if page is actually not deleted
			is404, err := aggregator.Is404()
			if err != nil {
				activeTask.Err = err
				markCompleted()
				return
			}
			if is404 {
				activeTask.Err = fmt.Errorf(
					"Aggregator: The requested page has been taken down or is invalid",
				)
				markCompleted()
				return
			}

			// evaluate displayName filename
			fname, err := aggregator.EvaluateFileName()
			if fname != "" {
				activeTask.DisplayName = fname
			}

			// get download page
			dlPage, err := aggregator.EvaluateDownloadPage()
			if err != nil {
				activeTask.Err = err
				markCompleted()
				return
			}
			defer dlPage.Close()

			// parse a filehost downloader
			filehostConstructor, err := engine.EvaluateFilehost(dlPage.URL())
			if err != nil {
				activeTask.Err = err
				markCompleted()
				return
			}
			filehost := filehostConstructor(dlPage)

			activeTask.FilehostUrl = filehost.Page().URL()

			// evaluate final filename
			if fname == "" {
				fname, err = filehost.EvaluateFileName()
				if err != nil {
					activeTask.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the filename")
					markCompleted()
					return
				}

				// setting the filename only if it is stil not set
				activeTask.DisplayName = fname
			}

			fext, err := aggregator.EvaluateFileExt()
			if err != nil {
				fext, err = filehost.EvaluateFileExt()
				if err != nil {

					activeTask.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the file extension")
					markCompleted()
					return
				}
			}

			// re-check if task is already done by other means
			found, _, _ := tq.Find(activeTask)
			if found {
				activeTask.SetErrMsg("This task is already present in the database")
				markCompleted()
				return
			}

			// check if out dirs exist
			if !appUtils.DirectoryExists(downloadDir) {
				err := appUtils.MkdirAll(downloadDir)
				if err != nil {
					log.Fatalln("taskRunner.DirCheck:", err)
				}
			}

			if !appUtils.DirectoryExists(tempDir) {
				err := appUtils.MkdirAll(tempDir)
				if err != nil {
					log.Fatalln("taskRunner.DirCheck:", err)
				}
			}

			// download the file into temp
			fullFilename := fmt.Sprintf("%s.%s", fname, fext)

			updateHandler := func(prog int8) {
				activeTask.SetProgress(prog)

				publisher.Publish(&pubsub.PublishEvent{
					EvtType: "update-node-content",
					Data:    activeTask,
				})
			}

			err = filehost.Download(tempDir, downloadDir, fullFilename, updateHandler)
			if err != nil {
				activeTask.SetErr(err)
				markCompleted()
				return
			}

			// task done :)
			markCompleted()
		}
	}
}
