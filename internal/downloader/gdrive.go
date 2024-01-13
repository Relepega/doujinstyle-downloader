package downloader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

func querySelectorVal(p playwright.Page, eval string) (string, error) {
	valInterface, err := p.Evaluate(eval)
	if err != nil {
		return "", err
	}

	val, _ := valInterface.(string)

	return val, nil
}

func craftDirectDownloadLink(p playwright.Page) (string, error) {
	var id string
	var export string
	var confirm string
	var uuid string
	var err error

	id, err = querySelectorVal(p, `document.querySelector('input[name="id"]').value`)
	if err != nil {
		return "", err
	}

	export, err = querySelectorVal(p, `document.querySelector('input[name="export"]').value`)
	if err != nil {
		return "", err
	}

	confirm, err = querySelectorVal(p, `document.querySelector('input[name="confirm"]').value`)
	if err != nil {
		return "", err
	}

	uuid, err = querySelectorVal(p, `document.querySelector('input[name="uuid"]').value`)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintln(
		"https://drive.usercontent.google.com/download?id=" + id + "&export=" + export + "&confirm=" + confirm + "&uuid=" + uuid,
	)

	url = strings.TrimSpace(url)

	return url, nil
}

func GDrive(albumName string, dlPage playwright.Page, progress *int8) error {
	defer dlPage.Close()

	pageUrl := dlPage.URL()

	_, err := dlPage.Goto(
		"https://drive.google.com/u/0/uc?id=" + strings.Split(pageUrl, "/")[5] + "&export=download",
	)
	if err != nil {
		return err
	}

	res, err := dlPage.Evaluate(
		"document.querySelector('a').innerText.split('.').toReversed()[0]",
	)
	if err != nil {
		return err
	}

	extension := fmt.Sprintf(".%v", res)

	appConfig, err := configManager.NewConfig()
	if err != nil {
		return err
	}
	DOWNLOAD_ROOT := appConfig.Download.Directory

	fp := filepath.Join(DOWNLOAD_ROOT, albumName+extension)
	fileExists, _ := appUtils.FileExists(fp)
	if fileExists {
		return nil
	}

	err = dlPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	if err != nil {
		return err
	}

	dlUrl, err := craftDirectDownloadLink(dlPage)
	if err != nil {
		return err
	}

	err = appUtils.DownloadFile(fp, dlUrl, progress)
	if err != nil {
		return err
	}

	return nil
}
