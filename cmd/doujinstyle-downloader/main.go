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
	defer log.Println("---------- SESSION END ----------")

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
	defer func() {
		err := engine.Shutdown()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	server := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, engine)
	server.Start()

	// create channel that waits for a SIGTERM event
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer server.Shutdown(ctx)
}
