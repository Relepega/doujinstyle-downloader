package hosts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

func Mega(albumName string, dlPage playwright.Page, progress *int8) error {
	defer dlPage.Close()

	for {
		val, _ := dlPage.Evaluate(
			"() => document.querySelector('#loading').classList.contains('hidden')",
		)

		hidden, _ := val.(bool)
		if hidden {
			break
		}

		// if err == nil {
		// 	break
		// }

		// empty folder
		val, err := dlPage.Evaluate(
			"document.querySelectorAll('.fm-empty-cloud-txt')[2].innerText",
		)
		if err != nil && val != nil {
			return fmt.Errorf("mega: %v", val)
		}

		time.Sleep(time.Second * 5)
	}

	ext, err := dlPage.Evaluate("document.querySelector('.extension').innerText")
	if err != nil {
		ext = ".zip"
	}
	extension := fmt.Sprintf("%v", ext)

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

	timeout := 0.0
	downloadHandler, err := dlPage.ExpectDownload(func() error {
		_, err := dlPage.Evaluate(`() => {
			const selectors = ['.js-default-download', '.fm-download-as-zip']
			selectors.forEach(sel => {
				let el = document.querySelector(sel)
				if (el) {
					el.click()
					return
				}
			})
		}`)

		errorDiv := dlPage.Locator(".default-warning > .txt")

		re := regexp.MustCompile(`\d`)

		for {
			val, _ := dlPage.Evaluate(
				"() => document.querySelector('.transfer-task-status').innerText",
			)

			strVal, _ := val.(string)

			if strVal == "Completed" {
				break
			}

			visible, _ := errorDiv.IsVisible()
			if visible {
				errVal, _ := errorDiv.InnerText()
				return fmt.Errorf("%v", errVal)
			}

			// Empty folder
			msg, _ := dlPage.Evaluate(
				"document.querySelector('.mega-dialog.warning > header > .info-container > .text').innerText",
			)
			msgVal, _ := msg.(string)
			if msgVal != "" {
				return fmt.Errorf("Mega: %s", msgVal)
			}

			// Folder too big to download within the browser
			msg, _ = dlPage.Evaluate(
				"document.querySelector('.mega-dialog.confirmation > header > .info-container > #msgDialog-title').innerText",
			)
			msgVal, _ = msg.(string)
			if msgVal != "" {
				return fmt.Errorf("Mega: %s", msgVal)
			}

			// get download percentage
			pageTitle, _ := dlPage.Evaluate("document.title")
			pageTitleStr, ok := pageTitle.(string)
			if ok {
				match := re.FindString(pageTitleStr)
				conv, err := strconv.ParseInt(match, 10, 8)
				if err == nil {
					*progress = int8(conv)
				}
			}

			time.Sleep(time.Second)
		}

		return err
	}, playwright.PageExpectDownloadOptions{
		Timeout: &timeout,
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
		return err
	}

	return nil
}
