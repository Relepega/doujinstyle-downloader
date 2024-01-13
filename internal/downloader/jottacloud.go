package downloader

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

func Jottacloud(albumName string, dlPage playwright.Page, progress *int8) error {
	defer dlPage.Close()

	fnSelector := "[data-testid=FileViewerHeaderFileName]"

	for {
		res, err := dlPage.Evaluate(
			"() => document.querySelector('" + fnSelector + "')",
		)
		if err != nil {
			return err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	res, err := dlPage.Evaluate(
		"document.querySelector('" + fnSelector + "').childNodes[0].textContent.split('.')[1]",
	)
	if err != nil {
		return err
	}

	extension := fmt.Sprintf(".%v", res)

	appConfig, err := configManager.NewConfig()
	if err != nil {
		return err
	}
	DOWNLOAD_ROOT := appConfig.Download.Directory

	fp := filepath.Join(DOWNLOAD_ROOT, albumName+extension)
	fileExists, err := appUtils.FileExists(fp)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	href, err := dlPage.Evaluate("document.querySelector(\"a[download]\").href")
	if err != nil {
		return err
	}
	downloadUrl, ok := href.(string)
	if !ok {
		return fmt.Errorf("Jottacloud: Couldn't get download url")
	}

	err = appUtils.DownloadFile(fp, downloadUrl, progress)
	if err != nil {
		return err
	}

	return nil
}
