package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

func Mega(albumName string, dlPage *playwright.Page) error {
	for {
		val, _ := (*dlPage).Evaluate(
			"() => document.querySelector('#loading').classList.contains('hidden')",
		)

		hidden, _ := val.(bool)
		if hidden {
			break
		}

		val = nil

		// if err == nil {
		// 	break
		// }

		// empty folder
		val, err := (*dlPage).Evaluate(
			"document.querySelectorAll('.fm-empty-cloud-txt')[2].innerText",
		)
		if err != nil {
			return fmt.Errorf("mega: %v", val)
		}
		val = nil

		time.Sleep(time.Second * 5)
	}

	ext, err := (*dlPage).Evaluate("document.querySelector('.extension').innerText")
	if err != nil {
		err = nil
		ext = ".zip"
	}
	extension := fmt.Sprintf("%v", ext)

	fp := filepath.Join(DOWNLOAD_ROOT, albumName+extension)
	_, err = os.Stat(fp)
	if err == nil {
		return nil
	}

	timeout := 0.0
	downloadHandler, err := (*dlPage).ExpectDownload(func() error {
		_, err := (*dlPage).Evaluate(`() => {
			const selectors = ['.js-default-download', '.fm-download-as-zip']
			selectors.forEach(sel => {
				let el = document.querySelector(sel)
				if (el) {
					el.click()
					return
				}
			})
		}`)

		errorDiv := (*dlPage).Locator(".default-warning > .txt")

		for {
			val, _ := (*dlPage).Evaluate(
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
			msg, _ := (*dlPage).Evaluate(
				"document.querySelector('.mega-dialog.warning > header > .info-container > .text').innerText",
			)
			msgVal, _ := msg.(string)
			if msgVal != "" {
				return fmt.Errorf("Mega: %s", msgVal)
			}
			msg = nil
			msgVal = ""

			// Folder too big to download withing the browser
			msg, _ = (*dlPage).Evaluate(
				"document.querySelector('.mega-dialog.confirmation > header > .info-container > #msgDialog-title').innerText",
			)
			msgVal, _ = msg.(string)
			if msgVal != "" {
				return fmt.Errorf("Mega: %s", msgVal)
			}

			time.Sleep(time.Second)
		}

		return err
	}, playwright.PageExpectDownloadOptions{
		Timeout: &timeout,
	})
	if err != nil {
		fmt.Println("Error expecting download:", err)
		return err
	}

	err = (*dlPage).Close()
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
