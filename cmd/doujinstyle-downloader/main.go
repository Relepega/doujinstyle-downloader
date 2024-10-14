package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/logger"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	"github.com/relepega/doujinstyle-downloader/internal/store"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue"
	"github.com/relepega/doujinstyle-downloader/internal/webserver"
)

func main() {
	defer log.Println("---------- SESSION END ----------")

	// init logger
	logdir := filepath.Join(".", "Logs")
	err := appUtils.CreateFolder(logdir)
	if err != nil {
		log.Fatalln(err)
	}

	logger.InitLogger(logdir)

	log.Println("---------- SESSION START ----------")

	// install playwright browsers
	err = playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium", "firefox"},
	})
	if err != nil {
		log.Fatalf("Couldn't install playwright dependencies: %v", err)
	}

	fmt.Print("\033[H\033[2J")

	// init config
	cfg := configManager.NewConfig()

	err = cfg.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}

	err = cfg.Save()
	if err != nil {
		log.Fatal(err)
	}

	// set config to the store value
	store.GetStore().Set("app-config", cfg)

	// create download folder
	err = appUtils.CreateFolder(cfg.Download.Directory)
	if err != nil {
		log.Fatal(err)
	}

	// "clean" temp dir
	tempdir := cfg.App.Tempdir
	appUtils.SetTempDir(tempdir)

	exists, err := appUtils.DirectoryExists(tempdir)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		err = os.RemoveAll(tempdir)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = appUtils.CreateAppTempDir(tempdir)
	if err != nil {
		log.Fatal(err)
	}

	// Init new default event publisher
	pub := pubsub.NewGlobalPublisher("sse")

	// init context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// init playwright
	pwc, err := playwrightWrapper.UsePlaywright("firefox", !cfg.Dev.PlaywrightDebug, 0.0)
	defer func() {
		_ = pwc.Close()
	}()

	// init and run queue
	q := taskQueue.NewQueue(cfg.Download.ConcurrentJobs, pub)

	go func() {
		if err != nil {
			log.Fatal(err)
		}

		q.Run(ctx, pwc)
	}()

	// init and run webserver
	webserver := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, q)

	go func() {
		err := webserver.Start(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// graceful shutdown
	<-ctx.Done()

	log.Println("App shut down successfully")
}
