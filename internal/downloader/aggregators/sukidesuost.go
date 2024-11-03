package aggregators

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

const (
	SDO_HOSTNAME         = "sukidesuost.info"
	SDO_ALBUM_URL        = "https://" + SDO_HOSTNAME + "/"
	SDO_INVALID_TYPE_ERR = "value is not a string:"
)

type SukiDesuOST struct {
	dsdl.Aggregator

	url  string
	page playwright.Page
}

func NewSukiDesuOst(slug string, p playwright.Page) dsdl.AggregatorImpl {
	var url string

	if strings.Contains(slug, SDO_HOSTNAME) {
		if strings.HasPrefix(slug, "http") {
			url = slug
		} else {
			url = "https://" + slug
		}
	} else {
		url = SDO_ALBUM_URL + slug
	}

	return &SukiDesuOST{
		page: p,
		url:  url,
	}
}

func (s *SukiDesuOST) Url() string {
	return s.url
}

func (s *SukiDesuOST) Page() playwright.Page {
	return s.page
}

func (sdo *SukiDesuOST) Is404() (bool, error) {
	valInterface, _ := sdo.page.Evaluate(
		"() => document.querySelector('.jeg_404_content') ? true : false",
	)
	val, ok := valInterface.(bool)

	if !ok {
		return false, fmt.Errorf("Could not convert value: %v", val)
	}

	if val {
		return true, nil
	}

	return false, nil
}

func (sdo *SukiDesuOST) EvaluateFileName() (string, error) {
	valInterface, err := sdo.page.Evaluate("document.querySelector('.jeg_post_title').innerText")
	if err != nil {
		return "", err
	}

	fn, ok := valInterface.(string)
	if !ok {
		return "", fmt.Errorf("%s %v", SDO_INVALID_TYPE_ERR, fn)
	}

	filename := appUtils.SanitizePath(strings.ReplaceAll(fn, " – ", " — "))
	filename = appUtils.SanitizePath(strings.ReplaceAll(filename, " - ", " — "))

	audioFormatsInterface, err := sdo.page.Evaluate(
		"document.querySelector('.content-inner > p:nth-child(3)').childNodes[0].data.split(': ')[1]",
	)
	if err != nil {
		return filename, nil
	}

	audiofmts, ok := audioFormatsInterface.(string)
	if !ok {
		return filename, nil
	}

	audiofmts = appUtils.SanitizePath(strings.ReplaceAll(audiofmts, " - ", " — "))
	filename = fmt.Sprintf("%s [%s]", filename, audiofmts)

	return filename, nil
}

func (sdo *SukiDesuOST) EvaluateFileExt() (string, error) {
	return "", fmt.Errorf(dsdl.AGGR_ERR_UNAVAILABLE_FT)
}

func (sdo *SukiDesuOST) EvaluateDownloadPage() (playwright.Page, error) {
redoIfInvalid:
	jsSelectors := []string{
		"document.querySelector('.content-inner > ul > li > a').href",
		"document.querySelector('.entry-content > ul > li > a').href",
		"document.querySelector('.content-inner > p:nth-child(5) > span > a').href",
		// flac
		"document.querySelectorAll('.content-inner > p:nth-child(4) > a')[0].href",
		"document.querySelectorAll('.content-inner > p:nth-child(5) > a')[0].href",
		"document.querySelector('tr:nth-child(4) > td:nth-child(2) > strong > span > span > span > a').href",
		// mp3
		"document.querySelectorAll('.content-inner > p:nth-child(4) > a')[1].href",
		"document.querySelectorAll('.content-inner > p:nth-child(5) > a')[1].href",
		"document.querySelector('tr:nth-child(5) > td:nth-child(2) > strong > span > span > span > a').href",
	}

	dlUrl := ""

	for _, selector := range jsSelectors {
		dlUrlInterface, err := sdo.page.Evaluate(selector)
		if err != nil {
			continue
		}

		tempDlUrl, ok := dlUrlInterface.(string)
		if !ok {
			continue
		}

		if tempDlUrl != "" {
			dlUrl = tempDlUrl
			break
		}
	}

	if dlUrl == "" {
		return nil, fmt.Errorf("Couldn't get a download URL")
	}

	if strings.Contains(dlUrl, "cuty.io") {
		_, _ = sdo.page.Reload()
		goto redoIfInvalid
	}

	dlPage, err := sdo.page.Context().NewPage()
	if err != nil {
		return nil, err
	}

	_, err = dlPage.Goto(dlUrl, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, err
	}

	return dlPage, nil
}
