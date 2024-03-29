package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/hosts"
)

const (
	DOUJINSTYLE_ALBUM_URL       = "https://doujinstyle.com/?p=page&type=1&id="
	DEFAULT_PAGE_NOT_LOADED_ERR = "The download page did not load in a reasonable amount of time."
)

type doujinstyle struct {
	Service
	albumID  string
	bw       *playwright.Browser
	progress *int8
}

func (d *doujinstyle) checkDMCA(p *playwright.Page) (bool, error) {
	valInterface, _ := (*p).Evaluate(
		"() => document.querySelector('h3').innerText == 'Insufficient information to display content.'",
	)
	valBool, _ := valInterface.(bool)

	if valBool {
		return true, nil
	}

	return false, nil
}

func (d *doujinstyle) Process() error {
	ctx, err := (*d.bw).NewContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	_, err = page.Goto(DOUJINSTYLE_ALBUM_URL+d.albumID, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return err
	}

	isDMCA, err := d.checkDMCA(&page)
	if err != nil {
		return err
	}
	if isDMCA {
		return fmt.Errorf("Doujinstyle: %s", SERVICE_ERROR_404)
	}

	albumName, err := d.evaluateFilename(page)
	if err != nil {
		return err
	}

	dlPage, err := ctx.ExpectPage(func() error {
		_, err := page.Evaluate("document.querySelector('#downloadForm').click()")
		return err
	})
	if err != nil {
		return err
	}

	err = dlPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	if err != nil {
		runBeforeUnloadOpt := true

		pageCloseOptions := playwright.PageCloseOptions{
			RunBeforeUnload: &runBeforeUnloadOpt,
		}

		dlPage.Close(pageCloseOptions)

		return fmt.Errorf(DEFAULT_PAGE_NOT_LOADED_ERR)
	}

	dlPageUrl := dlPage.URL()

	hostDownloader, err := hosts.Switch(dlPageUrl)
	if err != nil {
		return err
	}

	err = hostDownloader(albumName, dlPage, d.progress)

	return err
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

func (d *doujinstyle) evaluateFilename(page playwright.Page) (string, error) {
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
