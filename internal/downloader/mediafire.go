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

func Mediafire(albumName string, dlPage playwright.Page) error {
	runBeforeUnloadOpt := true
	pageCloseOptions := playwright.PageCloseOptions{
		RunBeforeUnload: &runBeforeUnloadOpt,
	}

	timeout := 0.0

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

	defer func() {
		if dlPage != nil {
			dlPage.Close(pageCloseOptions)
		}
	}()

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

	downloadHandle, err := dlPage.ExpectDownload(func() error {
		_, err := dlPage.Evaluate("document.querySelector('#downloadButton').click()")
		if err != nil {
			return err
		}

		popupPage, _ := dlPage.ExpectPopup(func() error {
			return nil
		}, playwright.PageExpectPopupOptions{
			Timeout: &timeout,
		})
		if popupPage != nil {
			popupPage.Close(pageCloseOptions)
		}

		return nil
	}, playwright.PageExpectDownloadOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}

	err = dlPage.Close()
	if err != nil {
		return err
	}
	dlPage = nil

	err = downloadHandle.SaveAs(fp)
	if err != nil {
		return err
	}

	return nil
}
