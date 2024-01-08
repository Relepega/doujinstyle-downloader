package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/webserver"

	"github.com/playwright-community/playwright-go"
)

const dateFormatting = "2006-01-02"

func createLogFile(logdir string) (*os.File, error) {
	fn := fmt.Sprintf("log %v.log", time.Now().Format(dateFormatting))
	fp := filepath.Join(logdir, fn)

	file, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error creating log file: %v", err)
	}

	return file, nil
}

func createDownloadFolder() error {
	appConfig, err := configManager.NewConfig()
	if err != nil {
		return err
	}

	downloadRoot := appConfig.Download.Directory

	return appUtils.CreateFolder(downloadRoot)
}

func init() {
	logdir := filepath.Join(".", "Logs")
	err := appUtils.CreateFolder(logdir)
	if err != nil {
		log.Fatalln(err)
	}

	fileHandle, err := createLogFile(logdir)
	if err != nil {
		log.Fatalln(err)
	}

	logger := slog.New(slog.NewTextHandler(fileHandle, nil)) // or os.Stdout
	slog.SetDefault(logger)
}

func main() {
	log.Println("---------- SESSION START ----------")

	err := createDownloadFolder()
	if err != nil {
		log.Fatalln(err)
	}

	err = playwright.Install()
	if err != nil {
		log.Fatalf("Couldn't install playwright dependencies: %v", err)
	}

	webserver.StartWebserver()
}
