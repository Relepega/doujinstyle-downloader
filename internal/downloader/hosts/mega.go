package hosts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/appUtils"
	pubsub "github.com/relepega/doujinstyle-downloader-reloaded/internal/pubSub"
	tq_eventbroker "github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue/tq_event_broker"
)

type mega struct {
	Host

	page playwright.Page

	albumID   string
	albumName string

	dlPath     string
	dlProgress *int8
}

func newMega(p playwright.Page, albumID, albumName, downloadPath string, progress *int8) Host {
	return &mega{
		page: p,

		albumID:   albumID,
		albumName: albumName,

		dlPath:     downloadPath,
		dlProgress: progress,
	}
}

func (m *mega) Download() error {
	for {
		val, _ := m.page.Evaluate(
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
		val, err := m.page.Evaluate(
			"document.querySelectorAll('.fm-empty-cloud-txt')[2].innerText",
		)
		if err != nil && val != nil {
			return fmt.Errorf("mega: %v", val)
		}

		time.Sleep(time.Second * 5)
	}

	ext, err := m.page.Evaluate("document.querySelector('.extension').innerText")
	if err != nil {
		ext = ".zip"
	}
	extension := fmt.Sprintf("%v", ext)

	fp := filepath.Join(m.dlPath, m.albumName+extension)
	fileExists, _ := appUtils.FileExists(fp)
	if fileExists {
		return nil
	}

	timeout := 0.0
	downloadHandler, err := m.page.ExpectDownload(func() error {
		_, err := m.page.Evaluate(`() => {
			const selectors = ['.js-default-download', '.fm-download-as-zip']
			selectors.forEach(sel => {
				let el = document.querySelector(sel)
				if (el) {
					el.click()
					return
				}
			})
		}`)

		errorDiv := m.page.Locator(".default-warning > .txt")

		re := regexp.MustCompile(`\d`)

		for {
			val, _ := m.page.Evaluate(
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
			msg, _ := m.page.Evaluate(
				"document.querySelector('.mega-dialog.warning > header > .info-container > .text').innerText",
			)
			msgVal, _ := msg.(string)
			if msgVal != "" {
				return fmt.Errorf("Mega: %s", msgVal)
			}

			// Folder too big to download within the browser
			msg, _ = m.page.Evaluate(
				"document.querySelector('.mega-dialog.confirmation > header > .info-container > #msgDialog-title').innerText",
			)
			msgVal, _ = msg.(string)
			if msgVal != "" {
				return fmt.Errorf("Mega: %s", msgVal)
			}

			pub, _ := pubsub.GetGlobalPublisher("queue")

			// get download percentage
			pageTitle, _ := m.page.Evaluate("document.title")
			pageTitleStr, ok := pageTitle.(string)
			if ok {
				match := re.FindString(pageTitleStr)
				conv, err := strconv.ParseInt(match, 10, 8)
				if err == nil {
					*m.dlProgress = int8(conv)

					pub.Publish(&pubsub.PublishEvent{
						EvtType: "update-task-progress",
						Data: &tq_eventbroker.UpdateTaskProgress{
							Id:       m.albumID,
							Progress: *m.dlProgress,
						},
					})
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

	err = m.page.Close()
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
