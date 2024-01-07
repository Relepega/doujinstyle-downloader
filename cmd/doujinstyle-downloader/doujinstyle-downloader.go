package main

import (
	"log"

	"github.com/relepega/doujinstyle-downloader/internal/webserver"

	"github.com/playwright-community/playwright-go"
)

func main() {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("Couldn't install playwright dependencies: %v", err)
	}

	webserver.StartWebserver()
}
