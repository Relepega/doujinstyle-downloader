package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

func Mega(albumName string, dlPage *playwright.Page) error {
	// locator := (*dlPage).Locator(".js-default-download")
	// err := locator.WaitFor(playwright.LocatorWaitForOptions{
	// 	State: playwright.WaitForSelectorStateVisible,
	// })
	// if err != nil {
	// 	fmt.Println("here")
	// 	locator = (*dlPage).Locator(".mega-button.fm-download-as-zip")
	// 	err = locator.WaitFor(playwright.LocatorWaitForOptions{
	// 		State: playwright.WaitForSelectorStateVisible,
	// 	})
	//
	// 	if err != nil {
	// 		return fmt.Errorf(
	// 			"No download button found. Required manual intervention.\nLog: %v",
	// 			err,
	// 		)
	// 	}
	// }

	for {
		val, _ := (*dlPage).Evaluate(
			"() => document.querySelector('#loading').classList.contains('hidden')",
		)

		hidden, _ := val.(bool)
		if hidden {
			break
		}

		// if err == nil {
		// 	break
		// }

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
		// return (*dlPage).Locator(".js-default-download").Click()
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

			time.Sleep(time.Second)
		}

		return err
	}, playwright.PageExpectDownloadOptions{
		Timeout: &timeout,
	})
	if err != nil {
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
