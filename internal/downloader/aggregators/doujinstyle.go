package aggregators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

const (
	DOUJINSTYLE_ALBUM_URL       = "https://doujinstyle.com/?p=page&type=1&id="
	DEFAULT_PAGE_NOT_LOADED_ERR = "The download page did not load in a reasonable amount of time."
)

type Doujinstyle struct {
	dsdl.Aggregator

	url  string
	page playwright.Page
}

func NewDoujinstyle(slug string, p playwright.Page) dsdl.AggregatorImpl {
	var url string

	if strings.HasSuffix(slug, "https") {
		url = slug
	} else {
		url = fmt.Sprintf("%s%s", DOUJINSTYLE_ALBUM_URL, slug)
	}

	return &Doujinstyle{
		page: p,
		url:  url,
	}
}

func (d *Doujinstyle) Url() string {
	return d.url
}

func (d *Doujinstyle) Page() playwright.Page {
	return d.page
}

func (d *Doujinstyle) Is404() (bool, error) {
	valInterface, err := d.page.Evaluate(
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

func (d *Doujinstyle) EvaluateFileName() (string, error) {
	album, err := d.page.Evaluate("document.querySelector('h2').innerText")
	if err != nil {
		return "", err
	}

	artist, err := d.page.Evaluate("document.querySelectorAll('.pageSpan2')[0].innerText")
	if err != nil {
		return "", err
	}

	format, err := d.page.Evaluate(`
	   Array.from(document.querySelectorAll(".d.page.pan1")).find(el => el.innerText == "Format:").nextElementSibling.innerText
	`)
	if err != nil {
		return "", err
	}

	val, err := d.page.Evaluate("document.querySelectorAll('.pageSpan2')[1].innerText")
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

func (d *Doujinstyle) EvaluateFileExt() (string, error) {
	return "", fmt.Errorf(dsdl.AGGR_ERR_UNAVAILABLE_FT)
}

func (d *Doujinstyle) EvaluateDownloadPage() (playwright.Page, error) {
	dlPage, err := d.page.Context().ExpectPage(func() error {
		_, err := d.page.Evaluate("document.querySelector('#downloadForm').click()")
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

/*

aggregator-specific functions

*/

func (d *Doujinstyle) getExhibitions(strVal string) string {
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
