package main

import (
	"errors"
	"log"
	"os"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
)

func initConfig() *configManager.Config {
	cfg := configManager.NewConfig()

	err := cfg.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}

	err = cfg.Save()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
