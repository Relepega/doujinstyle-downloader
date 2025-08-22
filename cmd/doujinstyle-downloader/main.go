package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/initters"
	"github.com/relepega/doujinstyle-downloader/internal/logger"
	webserver "github.com/relepega/doujinstyle-downloader/internal/webserver"
)

func main() {
	// init logger
	logdir := filepath.Join(".", "Logs")
	err := appUtils.MkdirAll(logdir)
	if err != nil {
		log.Fatalln(err)
	}

	logger.InitLogger(logdir)

	log.Println("--------- SESSION START ---------")

	// install playwright browsers
	err = playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium", "firefox"},
	})
	if err != nil {
		log.Fatalf("Couldn't install playwright dependencies: %v", err)
	}

	// clear screen
	// fmt.Print("\033[H\033[2J")

	// init modules
	cfg := initters.InitConfig()

	engine := initters.InitEngine(cfg)

	server := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, engine)
	server.Start()

	// engine runner
	stopRunner := make(chan struct{})
	go func(engine *dsdl.DSDL, stopRunner chan struct{}) {
		initters.QueueRunner(engine, cfg, stopRunner)
	}(engine, stopRunner)

	// create channel that waits for a SIGTERM event
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Main: Termination signal caught")

	log.Println("Main: Stopping QueueRunner")
	stopRunner <- struct{}{}
	close(stopRunner)

	log.Println("Main: Creating shutdown context")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Main: Shutting down webserver")
	server.Shutdown(ctx)

	log.Println("Main: Shutting down engine")
	err = engine.Shutdown()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Main: All modules have been shut down, closing the app")
	log.Println("---------- SESSION END ----------")
}
