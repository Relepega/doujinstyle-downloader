package downloader

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/configManager"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/downloader/services"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/playwrightWrapper"
)

func Download(serviceName, albumID string, pwc *playwrightWrapper.PwContainer) error {
	runBeforeUnloadOpt := true
	pageCloseOpts := playwright.PageCloseOptions{
		RunBeforeUnload: &runBeforeUnloadOpt,
	}

	cfg := configManager.NewConfig()
	err := cfg.Load()
	if err != nil {
		return err
	}

	ctx, err := pwc.Browser.NewContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	service, err := services.NewService(serviceName, albumID)
	if err != nil {
		return err
	}

	servicePage, err := service.OpenServicePage(&ctx)
	if err != nil {
		return err
	}
	defer servicePage.Close(pageCloseOpts)

	isDMCA, err := service.CheckDMCA(&servicePage)
	if err != nil {
		return err
	}

	if isDMCA {
		return fmt.Errorf("Doujinstyle: %s", services.SERVICE_ERROR_404)
	}

	mediaName, err := service.EvaluateFilename(&servicePage)
	if err != nil {
		return err
	}

	downloadPage, err := service.OpenDownloadPage(servicePage)
	if err != nil {
		return err
	}
	defer downloadPage.Close()

	_ = servicePage.Close(pageCloseOpts)

	return nil
}
