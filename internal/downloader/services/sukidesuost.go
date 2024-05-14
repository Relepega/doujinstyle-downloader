package services

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
)

const (
	SDO_ALBUM_URL        = "https://sukidesuost.info/"
	SDO_INVALID_TYPE_ERR = "value is not a string:"
)

type sukidesuost struct {
	Service

	urlSlug string
}

func newSukidesuost(mediaID string) Service {
	return &sukidesuost{
		urlSlug: mediaID,
	}
}

func (sdo *sukidesuost) OpenServicePage(ctx *playwright.BrowserContext) (playwright.Page, error) {
	p, err := (*ctx).NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	_, err = p.Goto(SDO_ALBUM_URL+sdo.urlSlug, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, fmt.Errorf("could not open sukidesuost page: %v", err)
	}

	return p, nil
}

func (sdo *sukidesuost) CheckDMCA(p playwright.Page) (bool, error) {
	valInterface, _ := p.Evaluate(
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

func (sdo *sukidesuost) EvaluateFilename(p playwright.Page) (string, error) {
	valInterface, err := p.Evaluate("document.querySelector('.jeg_post_title').innerText")
	if err != nil {
		return "", err
	}

	fn, ok := valInterface.(string)
	if !ok {
		return "", fmt.Errorf("%s %v", SDO_INVALID_TYPE_ERR, fn)
	}

	filename := appUtils.SanitizePath(strings.ReplaceAll(fn, " - ", " — "))

	audioFormatsInterface, err := p.Evaluate("document.querySelector('.jeg_post_title').innerText")
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

func (sdo *sukidesuost) OpenDownloadPage(servicePage playwright.Page) (playwright.Page, error) {
	jsSelectors := []string{
		"document.querySelector('.content-inner > ul > li > a').href",
		"document.querySelectorAll('.content-inner > p:nth-child(4) > a')[0].href",
		"document.querySelectorAll('.content-inner > p:nth-child(4) > a')[1].href",
	}

	var dlUrl string

	for _, selector := range jsSelectors {
		dlUrlInterface, err := servicePage.Evaluate(selector)
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

	// dlUrlInterface, err := servicePage.Evaluate(
	// 	"document.querySelector('.content-inner > ul > li > a').href",
	// )
	// if err != nil {
	// 	return nil, err
	// }
	//
	// dlUrl, ok := dlUrlInterface.(string)
	// if !ok {
	// 	return nil, fmt.Errorf("%s %v", SDO_INVALID_TYPE_ERR, dlUrl)
	// }

	dlPage, err := servicePage.Context().NewPage()
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
