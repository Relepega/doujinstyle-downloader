package services

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/hosts"
)

const (
	SDO_INVALID_TYPE_ERR = "value is not a string:"
	SDO_ALBUM_URL        = "https://sukidesuost.info/"
)

type sukiDesuOst struct {
	Service
	urlSlug  string
	bw       *playwright.Browser
	progress *int8
}

func (sdo *sukiDesuOst) checkDMCA(p *playwright.Page) (bool, error) {
	valInterface, _ := (*p).Evaluate(
		"() => document.querySelector('.jeg_404_content') ? true : false",
	)
	val, _ := valInterface.(bool)

	if val {
		return true, nil
	}

	return false, nil
}

func (sdo *sukiDesuOst) evaluateFilename(p playwright.Page) (string, error) {
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

func (sdo *sukiDesuOst) Process() error {
	ctx, err := (*sdo.bw).NewContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	_, err = page.Goto(SDO_ALBUM_URL+sdo.urlSlug, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return err
	}

	isDMCA, err := sdo.checkDMCA(&page)
	if err != nil {
		return err
	}
	if isDMCA {
		return fmt.Errorf("Sukidesuost: %s", SERVICE_ERROR_404)
	}

	albumName, err := sdo.evaluateFilename(page)
	if err != nil {
		return err
	}

	dlUrlInterface, err := page.Evaluate(
		"document.querySelector('.content-inner > ul > li > a').href",
	)
	if err != nil {
		return err
	}

	dlUrl, ok := dlUrlInterface.(string)
	if !ok {
		return fmt.Errorf("%s %v", SDO_INVALID_TYPE_ERR, dlUrl)
	}

	dlPage, err := ctx.NewPage()
	if err != nil {
		return err
	}
	defer dlPage.Close()

	_, err = dlPage.Goto(dlUrl, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return err
	}

	hostDownloader, err := hosts.Switch(dlUrl)
	if err != nil {
		return err
	}

	err = hostDownloader(albumName, dlPage, sdo.progress)

	return err
}
