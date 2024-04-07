package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/relepega/doujinstyle-downloader-reloaded/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/configManager"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/logger"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/playwrightWrapper"
	pubsub "github.com/relepega/doujinstyle-downloader-reloaded/internal/pubSub"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver"

	"github.com/playwright-community/playwright-go"
)

func main() {
	// init logger
	logdir := filepath.Join(".", "Logs")
	err := appUtils.CreateFolder(logdir)
	if err != nil {
		log.Fatalln(err)
	}

	logger.InitLogger(logdir)

	log.Println("---------- SESSION START ----------")

	// install playwright browsers
	err = playwright.Install()
	if err != nil {
		log.Fatalf("Couldn't install playwright dependencies: %v", err)
	}

	// init config
	cfg := configManager.NewConfig()

	err = cfg.Load()
	if err != nil {
		err := cfg.Save()
		if err != nil {
			log.Fatal(err)
		}
	}

	// create download folder
	err = appUtils.CreateFolder(cfg.Download.Directory)
	if err != nil {
		log.Fatal(err)
	}

	// Init new default event publisher
	pub := pubsub.GetExistingPublisher()

	// init and run queue
	q := taskQueue.NewQueue(cfg.Download.ConcurrentJobs, pub)

	go func() {
		pwc, err := playwrightWrapper.UsePlaywright("chromium", !cfg.Dev.PlaywrightDebug, 0.0)
		if err != nil {
			log.Fatal(err)
		}

		q.Run(pwc)
	}()

	// init and run webserver
	webserver := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, q)

	go func() {
		webserver.Start()
	}()

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-ctx.Done()
	q.Quit <- nil
}
