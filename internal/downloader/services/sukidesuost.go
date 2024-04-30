package services

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/appUtils"
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

	return appUtils.SanitizePath(strings.ReplaceAll(fn, " - ", " — ")), nil
}

func (sdo *sukidesuost) OpenDownloadPage(servicePage playwright.Page) (playwright.Page, error) {
	dlUrlInterface, err := servicePage.Evaluate(
		"document.querySelector('.content-inner > ul > li > a').href",
	)
	if err != nil {
		return nil, err
	}

	dlUrl, ok := dlUrlInterface.(string)
	if !ok {
		return nil, fmt.Errorf("%s %v", SDO_INVALID_TYPE_ERR, dlUrl)
	}

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
