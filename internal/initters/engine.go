package initters

import (
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/aggregators"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/filehosts"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	"github.com/relepega/doujinstyle-downloader/internal/task"
)

func InitEngine(cfg *configManager.Config) *dsdl.DSDL {
	log.Println("Engine: Starting playwright")
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
	log.Println("Engine: Playwright started without errors")

	log.Println("Engine: Initializing DSDL instance")
	engine := dsdl.NewDSDL(pww.Browser)

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

	log.Println("Engine: DSDL initialized")

	return engine
}

func sendTaskUpdates(pub chan *pubsub.PublishEvent, t *task.Task) {
	pub <- &pubsub.PublishEvent{
		EvtType: "update-node-content",
		Data:    t,
	}
}

func QueueRunner(
	engine *dsdl.DSDL,
	cfg *configManager.Config,
	stop chan struct{},
) error {
	log.Println("QueueRunner: Starting running tasks")

	maxJobs := int(cfg.Download.ConcurrentJobs)

	db := engine.DB()

	var activeTasks []*task.Task

	for {
		select {
		case <-stop:
			log.Println("QueueRunner: Stopping runner and active tasks")
			for _, t := range activeTasks {
				t.Shutdown()
			}

			// make main wait for tasks to stop so that when closing the database we won't lose any data
			log.Println("QueueRunner: All tasks stopped")

			return nil

		default:
			if db.CountFromStateNoErr(states.TASK_STATE_QUEUED) <= 0 ||
				db.CountFromStateNoErr(states.TASK_STATE_RUNNING) == maxJobs {
				continue
			}

			log.Println("QueueRunner: Dequeuing task")

			t, err := db.GetFromState(states.TASK_STATE_QUEUED)
			if err != nil {
				continue
			}

			log.Printf("QueueRunner: Activating task with ID %v\n", t.Id)

			_, err = db.AdvanceState(t)
			if err != nil {
				continue
			}

			// swap any completed task if array is full
			if len(activeTasks) < maxJobs {
				activeTasks = append(activeTasks, t)
			} else {
				for i, v := range activeTasks {
					if v.DownloadState == states.TASK_STATE_COMPLETED {
						activeTasks[i] = t
						break
					}
				}
			}

			go taskRunner(engine, t, cfg.Download.Directory, cfg.Download.Tempdir)
		}
	}
}

func taskRunner(
	engine *dsdl.DSDL,
	t *task.Task,
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
		log.Printf("TaskRunner: Marking task %v as complete\n", t.Id)
		bwContext.Close()

		t.DownloadState = states.TASK_STATE_COMPLETED

		err := engine.DB().Update(t)
		if err != nil {
			log.Fatalf("TaskRunner: Error while updating task in DB: %v", err)
		}

		publisher.Publish(&pubsub.PublishEvent{
			EvtType: "mark-task-as-done",
			Data:    t,
		})
	}

	publisher.Publish(&pubsub.PublishEvent{
		EvtType: "activate-task",
		Data:    t,
	})

	running := false

	for {
		select {
		case msg := <-t.Stop:
			if msg == "user-abort" {
				t.Err = fmt.Errorf("Task aborted by user")
				markCompleted()

				return
			}

			if msg == "shutdown" {
				log.Printf("TaskRunner: Marking task as aborted (server shutdown) (ID: %v)\n", t.Id)

				err := bwContext.Close()
				log.Printf("TaskRunner: An error occurred while stopping task ID %v: %v", t.Id, err)

				engine.DB().Update(t)

				publisher.Publish(&pubsub.PublishEvent{
					EvtType: "mark-task-as-done",
					Data:    t,
				})

				return
			}

		default:
			if running {
				continue
			}

			// mark running, so that we don't end with a memory leak :)
			running = true

			// process the task
			aggConstFn, err := engine.EvaluateAggregator(t.Aggregator)
			if err != nil {
				t.Err = err
				markCompleted()
				return
			}

			bwContext, err = engine.Browser().NewContext()
			if err != nil {
				t.Err = fmt.Errorf("Playwright: Cannot open new browser context")
				markCompleted()
				return
			}
			defer bwContext.Close()

			p, err := bwContext.NewPage()
			if err != nil {
				t.Err = fmt.Errorf("Playwright: Cannot open new browser context page")
				markCompleted()
				return
			}
			defer p.Close()

			aggregator := aggConstFn(t.Slug, p)

			t.AggregatorPageURL = aggregator.Url()

			_, err = p.Goto(aggregator.Url())
			// check internet connection
			if err != nil {
				t.Err = err
				markCompleted()
				return
			}

			t.Slug = aggregator.Slug()

			// check if page is actually not deleted
			is404, err := aggregator.Is404()
			if err != nil {
				t.Err = err
				markCompleted()
				return
			}
			if is404 {
				t.Err = fmt.Errorf(
					"Aggregator: The requested page has been taken down or is invalid",
				)
				markCompleted()
				return
			}

			// evaluate displayName filename
			fname, err := aggregator.EvaluateFileName()
			if fname != "" {
				t.DisplayName = fname
			}

			publisher.Publish(&pubsub.PublishEvent{
				EvtType: "update-node-content",
				Data:    t,
			})
			engine.DB().Update(t)

			// get download page
			dlPage, err := aggregator.EvaluateDownloadPage()
			if err != nil {
				t.Err = err
				markCompleted()
				return
			}
			defer dlPage.Close()

			// parse a filehost downloader
			filehostConstructor, err := engine.EvaluateFilehost(dlPage.URL())
			if err != nil {
				t.Err = err
				markCompleted()
				return
			}
			filehost := filehostConstructor(dlPage)

			t.FilehostUrl = filehost.Page().URL()

			// evaluate final filename
			if fname == "" {
				fname, err = filehost.EvaluateFileName()
				if err != nil {
					t.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the filename")
					markCompleted()
					return
				}

				// setting the filename only if it is stil not set
				t.DisplayName = fname
			}

			fext, err := aggregator.EvaluateFileExt()
			if err != nil {
				fext, err = filehost.EvaluateFileExt()
				if err != nil {

					t.Err = fmt.Errorf("TaskRunner: Couldn't evaluate the file extension")
					markCompleted()
					return
				}
			}

			// re-check if task is already done by other means
			found, _, _ := engine.DB().Find(t.DisplayName)
			if found {
				t.SetErrMsg("This task is already present in the database")
				markCompleted()
				return
			}

			engine.DB().Update(t)
			publisher.Publish(&pubsub.PublishEvent{
				EvtType: "update-node-content",
				Data:    t,
			})

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
				t.SetProgress(prog)

				// fmt.Println("downloading (", prog, "%)", t.DisplayName)

				publisher.Publish(&pubsub.PublishEvent{
					EvtType: "update-node-content",
					Data:    t,
				})
			}

			err = filehost.Download(tempDir, downloadDir, fullFilename, updateHandler)
			if err != nil {
				t.SetErr(err)
				markCompleted()
				return
			}

			// task done :)
			markCompleted()
		}
	}
}
