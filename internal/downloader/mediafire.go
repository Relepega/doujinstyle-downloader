package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func Mediafire(albumName string, dlPage *playwright.Page) error {
	for {
		res, err := (*dlPage).Evaluate(
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

	ext, err := (*dlPage).Evaluate("document.querySelector('.filetype').innerText")
	if ext == nil {
		ext, err = (*dlPage).Evaluate(`() => {
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

	fp := filepath.Join(DOWNLOAD_ROOT, albumName+extension)
	_, err = os.Stat(fp)
	if err == nil {
		return nil
	}

	downloadHandler, err := (*dlPage).ExpectDownload(func() error {
		_, err := (*dlPage).Evaluate("document.querySelector('#downloadButton').click()")
		// err := dlPage.Locator("#downloadButton").Click()
		return err
	})
	if err != nil {
		return err
	}

	popupPage, popupErr := (*dlPage).ExpectPopup(func() error {
		return nil
	})

	if popupErr == nil {
		popupPage.Close()
	}

	err = (*dlPage).Close()
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
