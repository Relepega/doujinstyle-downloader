package downloader

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/services"
)

func Download(urlSlug string, bw *playwright.Browser, progress *int8) error {
	service := services.NewService(services.Doujinstyle, urlSlug, bw, progress)
	if service == nil {
		return fmt.Errorf("Not a valid service")
	}

	err := service.Process()
	if err != nil {
		return err
	}

	return nil
}
