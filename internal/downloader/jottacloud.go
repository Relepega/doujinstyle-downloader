package downloader

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

func Jottacloud(albumName string, dlPage playwright.Page) error {
	defer dlPage.Close()

	res, err := dlPage.Evaluate(
		"document.querySelector('[data-testid=FileViewerHeaderFileName]').childNodes[0].textContent.split('.')[1]",
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
	fileExists, _ := appUtils.FileExists(fp)
	if fileExists {
		return nil
	}

	downloadHandler, err := dlPage.ExpectDownload(func() error {
		_, err := dlPage.Evaluate("document.querySelector('.css-118jy9p.e16wmiuy0').click()")
		return err
	})
	if err != nil {
		return err
	}

	err = dlPage.Close()
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	err = downloadHandler.SaveAs(fp)
	if err != nil {
		return fmt.Errorf("%v\n--------------\n%v", err, downloadHandler.Failure())
	}

	return nil
}
