package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/logger"
)

func main() {
	defer log.Println("---------- SESSION END ----------")

	// init context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

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

	// init modules
	cfg := initConfig()
	engine := initEngine(ctx)
	server := initServer(cfg, ctx)
}
