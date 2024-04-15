package downloader

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/configManager"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/downloader/hosts"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/downloader/services"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/playwrightWrapper"
)

func Download(serviceName, albumID string, progress *int8, pwc *playwrightWrapper.PwContainer) error {
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
	defer servicePage.Close()

	isDMCA, err := service.CheckDMCA(servicePage)
	if err != nil {
		return err
	}

	if isDMCA {
		return fmt.Errorf("%s: %s", serviceName, services.SERVICE_ERROR_404)
	}

	mediaName, err := service.EvaluateFilename(servicePage)
	if err != nil {
		return err
	}

	downloadPage, err := service.OpenDownloadPage(servicePage)
	if err != nil {
		return err
	}
	defer downloadPage.Close(pageCloseOpts)

	_ = servicePage.Close(pageCloseOpts)

	hostFactory, err := hosts.NewHost(downloadPage.URL())
	if err != nil {
		return err
	}

	host := hostFactory(downloadPage, albumID, mediaName, cfg.Download.Directory, progress)
	err = host.Download()
	if err != nil {
		return err
	}

	return nil
}
