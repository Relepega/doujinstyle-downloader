package main

import (
	"context"

	"github.com/relepega/doujinstyle-downloader/internal/downloader/aggregators"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

func initEngine(ctx context.Context) *dsdl.DSDL {
	engine := dsdl.NewDSDL(ctx)

	engine.RegisterAggregator(&dsdl.Aggregator{
		Name:        "doujinstyle",
		Constructor: aggregators.NewDoujinstyle,
	})

	engine.RegisterAggregator(&dsdl.Aggregator{
		Name:        "sukidesuost",
		Constructor: aggregators.NewSukiDesuOst,
	})

	return engine
}
