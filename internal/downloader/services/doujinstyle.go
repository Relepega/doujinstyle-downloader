package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
)

const (
	DOUJINSTYLE_ALBUM_URL       = "https://doujinstyle.com/?p=page&type=1&id="
	DEFAULT_PAGE_NOT_LOADED_ERR = "The download page did not load in a reasonable amount of time."
)

type doujinstyle struct {
	Service

	mediaID string
}

func newDoujinstyle(mediaID string) Service {
	return &doujinstyle{
		mediaID: mediaID,
	}
}

func (d *doujinstyle) OpenServicePage(ctx *playwright.BrowserContext) (playwright.Page, error) {
	p, err := (*ctx).NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	_, err = p.Goto(DOUJINSTYLE_ALBUM_URL+d.mediaID, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, fmt.Errorf("could not open doujinstyle page: %v", err)
	}

	return p, nil
}

func (d *doujinstyle) CheckDMCA(p playwright.Page) (bool, error) {
	valInterface, err := p.Evaluate(
		"() => document.querySelector('h3').innerText == 'Insufficient information to display content.'",
	)
	if err != nil {
		return false, fmt.Errorf("Could not evaluate selector: %v", err)
	}

	valBool, ok := valInterface.(bool)
	if !ok {
		return false, fmt.Errorf("Could not convert value: %v", err)
	}

	if valBool {
		return true, nil
	}

	return false, nil
}

func (d *doujinstyle) getExhibitions(strVal string) string {
	re := regexp.MustCompile("^(C[0-9]+)|(M[0-9]-[0-9]+)|(AC[0-9])$")
	matches := []string{}

	for _, substr := range strings.Split(strVal, ", ") {
		if re.MatchString(substr) {
			matches = append(matches, substr)
		}
	}

	var fullStr string

	if len(matches) == 0 {
		fullStr = ""
	} else {
		fullStr = " [" + strings.Join(matches, ", ") + "]"
	}

	return fullStr
}

func (d *doujinstyle) EvaluateFilename(page playwright.Page) (string, error) {
	album, err := page.Evaluate("document.querySelector('h2').innerText")
	if err != nil {
		return "", err
	}

	artist, err := page.Evaluate("document.querySelectorAll('.pageSpan2')[0].innerText")
	if err != nil {
		return "", err
	}

	format, err := page.Evaluate(`
	   Array.from(document.querySelectorAll(".pageSpan1")).find(el => el.innerText == "Format:").nextElementSibling.innerText
	`)
	if err != nil {
		return "", err
	}

	val, err := page.Evaluate("document.querySelectorAll('.pageSpan2')[1].innerText")
	if err != nil {
		return "", err
	}
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string: %v", val)
	}
	event := d.getExhibitions(strVal)

	return appUtils.SanitizePath(fmt.Sprintf("%s — %s%s [%s]", artist, album, event, format)), nil
}

func (d *doujinstyle) OpenDownloadPage(p playwright.Page) (playwright.Page, error) {
	dlPage, err := p.Context().ExpectPage(func() error {
		_, err := p.Evaluate("document.querySelector('#downloadForm').click()")
		return err
	})
	if err != nil {
		return nil, err
	}

	err = dlPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %v", DEFAULT_PAGE_NOT_LOADED_ERR, err)
	}

	return dlPage, nil
}
