package downloader

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

func GDrive(albumName string, dlPage playwright.Page) error {
	defer dlPage.Close()

	pageUrl := dlPage.URL()

	_, err := dlPage.Goto(
		"https://drive.google.com/u/0/uc?id=" + strings.Split(pageUrl, "/")[5] + "&export=download",
	)
	if err != nil {
		return err
	}

	res, err := dlPage.Evaluate(
		"document.querySelector('a').innerText.split('.').toReversed()[0]",
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
		_, err := dlPage.Evaluate("document.querySelector('#uc-download-link').click()")
		return err
	})
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
