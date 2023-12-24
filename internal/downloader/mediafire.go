package downloader

import (
	"fmt"
	"path/filepath"
	"regexp"
	"relepega/doujinstyle-downloader/internal/appUtils"
	"relepega/doujinstyle-downloader/internal/configManager"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func handlePopup(p playwright.Page) bool {
	p.Close()

	return false
}

func Mediafire(albumName string, dlPage playwright.Page) error {
	for {
		res, err := dlPage.Evaluate(
			"() => document.querySelector(\".DownloadStatus.DownloadStatus--uploading\")",
		)
		if err != nil {
			return err
		}

		if res == nil {
			break
		}

		time.Sleep(time.Second * 5)
	}

	var extension string

	defer dlPage.Close()

	ext, err := dlPage.Evaluate("document.querySelector('.filetype').innerText")
	if ext == nil {
		ext, _ = dlPage.Evaluate(`() => {
			let data = document.querySelector('.dl-btn-label').title.split('.')
			return data[data.length - 1]
		}`)

		extension = fmt.Sprintf(".%v", ext)
	} else {
		extension = fmt.Sprintf("%v", ext)

		re, err := regexp.Compile(`\.[a-zA-Z0-9]+`)
		if err != nil {
			return err
		}
		extension = strings.ToLower(re.FindString(extension))
	}

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

	download, err := dlPage.ExpectDownload(func() error {
		_, err = dlPage.Evaluate("document.querySelector('#downloadButton').click()")
		return err
	})
	if err != nil {
		err := download.Cancel()
		if err != nil {
			return err
		}

		return err
	}

	err = download.SaveAs(fp)
	if err != nil {
		return err
	}

	return nil
}
