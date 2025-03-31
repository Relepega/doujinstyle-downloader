package filehosts

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

type GDrive struct {
	dsdl.Filehost

	page playwright.Page
}

func NewGDrive(p playwright.Page) dsdl.FilehostImpl {
	return &GDrive{
		page: p,
	}
}

func (g *GDrive) SetPage(p playwright.Page) {
	g.page = p
}

func (g *GDrive) Page() playwright.Page {
	return g.page
}

func (g *GDrive) EvaluateFileName() (string, error) {
	// TODO
	return "", nil
}

func (g *GDrive) EvaluateFileExt() (string, error) {
	res, err := g.page.Evaluate(
		"document.querySelector('a').innerText.split('.').toReversed()[0]",
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", res), nil
}

func (g *GDrive) Download(tempDir, finalDir, filename string, setProgress func(p int8)) error {
	pageUrl := g.page.URL()

	_, err := g.page.Goto(
		"https://drive.google.com/u/0/uc?id=" + strings.Split(pageUrl, "/")[5] + "&export=download",
	)
	if err != nil {
		return err
	}

	err = g.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	if err != nil {
		return err
	}

	finalFilepath := filepath.Join(finalDir, filename)

	fileExists, err := appUtils.FileExists(finalFilepath)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	dlUrl, err := craftDirectDownloadLink(g.page)
	if err != nil {
		return err
	}

	err = appUtils.DownloadFile(dlUrl, tempDir, finalFilepath, setProgress)
	if err != nil {
		return err
	}

	return nil
}

/*

filehost-specific functions

*/

func craftDirectDownloadLink(p playwright.Page) (string, error) {
	querySelectorVal := func(eval string) (string, error) {
		valInterface, err := p.Evaluate(eval)
		if err != nil {
			return "", err
		}

		val, _ := valInterface.(string)

		return val, nil
	}

	var id string
	var export string
	var confirm string
	var uuid string
	var err error

	id, err = querySelectorVal(`document.querySelector('input[name="id"]').value`)
	if err != nil {
		return "", err
	}

	export, err = querySelectorVal(`document.querySelector('input[name="export"]').value`)
	if err != nil {
		return "", err
	}

	confirm, err = querySelectorVal(`document.querySelector('input[name="confirm"]').value`)
	if err != nil {
		return "", err
	}

	uuid, err = querySelectorVal(`document.querySelector('input[name="uuid"]').value`)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintln(
		"https://drive.usercontent.google.com/download?id=" + id + "&export=" + export + "&confirm=" + confirm + "&uuid=" + uuid,
	)

	url = strings.TrimSpace(url)

	return url, nil
}
