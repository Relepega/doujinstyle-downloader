package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/initters"
	"github.com/relepega/doujinstyle-downloader/internal/logger"
	webserver "github.com/relepega/doujinstyle-downloader/internal/webserver"
)

func main() {
	defer func() {
		log.Println("App shut down successfully")
		log.Println("---------- SESSION END ----------")
	}()

	// init context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// init logger
	logdir := filepath.Join(".", "Logs")
	err := appUtils.MkdirAll(logdir)
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

	// clear screen
	// fmt.Print("\033[H\033[2J")

	// init modules
	cfg := initters.InitConfig()

	engine := initters.InitEngine(cfg, ctx)
	defer engine.Tq.StopQueue()
	defer engine.GetBrowserInstance().Close()
	defer engine.Tq.GetDatabase().Close()

	server := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, ctx, engine)

	go func() {
		err := server.Start()
		if err != nil {
			log.Fatalln("Webserver:", err)
		}
	}()

	<-ctx.Done()
}
