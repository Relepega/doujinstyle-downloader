package filehosts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

type Mega struct {
	dsdl.Filehost

	page playwright.Page
}

func NewMega(p playwright.Page) dsdl.FilehostImpl {
	return &Mega{
		page: p,
	}
}

func (m *Mega) SetPage(p playwright.Page) {
	m.page = p
}

func (m *Mega) Page() playwright.Page {
	return m.page
}

func (m *Mega) waitForPageLoad() error {
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

	return nil
}

// TODO: implement this function
func (m *Mega) EvaluateFileName() (string, error) {
	err := m.waitForPageLoad()
	if err != nil {
		return "", err
	}

	return "", nil
}

func (m *Mega) EvaluateFileExt() (string, error) {
	err := m.waitForPageLoad()
	if err != nil {
		return "", err
	}

	ext, err := m.page.Evaluate("document.querySelector('.extension').innerText")
	if err != nil {
		// assume the file to download is a compressed folder.
		// by default mega compresses a folder in a .zip archive
		return "zip", nil
	}

	return fmt.Sprintf("%v", ext)[1:], nil
}

func (m *Mega) Download(tempDir, finalDir, filename string, setProgress func(p int8)) error {
	err := m.waitForPageLoad()
	if err != nil {
		return err
	}

	// waiting for loading spinner to disappear
	for {
		val, _ := m.page.Evaluate(
			"document.querySelectorAll('.loading-spinner')[2].classList.contains('hidden')",
		)
		isHidden, _ := val.(bool)

		if isHidden {
			time.Sleep(500 * time.Millisecond)
			break
		}
	}

	// limited quota
	// val, err := m.page.Evaluate(
	// 	"document.querySelector('.limited-bandwidth-dialog')?.getAttribute('aria-modal') == 'true'",
	// )
	// if err != nil && val == nil {
	// 	return err
	// }
	//
	// isLimited, _ := val.(bool)
	// if isLimited {
	// 	return fmt.Errorf("Mega: You’re running out of transfer quota")
	// }

	fp := filepath.Join(finalDir, filename)

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
		if err != nil {
			return fmt.Errorf("Mega: Couldn't start download: %v", err)
		}

		errorDiv := m.page.Locator(".default-warning > .txt")

		re := regexp.MustCompile(`\d+`)

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

			// get download percentage
			pageTitle, _ := m.page.Evaluate("document.title")
			pageTitleStr, ok := pageTitle.(string)
			if !ok {
				continue
			}

			match := re.FindString(pageTitleStr)
			conv, err := strconv.ParseInt(match, 10, 8)
			if err != nil {
				continue
			}

			setProgress(int8(conv))
		}

		return err
	}, playwright.PageExpectDownloadOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return err
	}

	err = downloadHandler.SaveAs(fp)
	if err != nil {
		return err
	}

	return nil
}
