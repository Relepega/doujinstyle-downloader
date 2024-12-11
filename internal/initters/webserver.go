package initters

import (
	"context"
	"log"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	webserver "github.com/relepega/doujinstyle-downloader/internal/webserver/v2"
)

func InitServer(
	cfg *configManager.Config,
	ctx context.Context,
	userData interface{},
) *webserver.Webserver {
	server := webserver.NewWebServer(cfg.Server.Host, cfg.Server.Port, ctx, userData)

	err := server.Start()
	if err != nil {
		log.Fatalln(err)
	}

	return server
}
