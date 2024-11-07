package main

import (
	"context"
	"log"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	webserver "github.com/relepega/doujinstyle-downloader/internal/webserver/v2"
)

func initServer(cfg *configManager.Config, ctx context.Context) *webserver.Webserver {
	server := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, ctx)

	err := server.Start(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return server
}
