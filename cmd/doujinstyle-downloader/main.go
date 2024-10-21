package main

import (
	"context"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

func main() {
	ctx := context.Background()

	d := dsdl.NewDSDL(ctx)
}
