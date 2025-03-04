package filehosts

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

type Jottacloud struct {
	dsdl.Filehost

	page playwright.Page
}

func NewJottacloud(p playwright.Page) dsdl.FilehostImpl {
	return &Jottacloud{
		page: p,
	}
}

func (j *Jottacloud) SetPage(p playwright.Page) {
	j.page = p
}

func (j *Jottacloud) Page() playwright.Page {
	return j.page
}

func (j *Jottacloud) EvaluateFileName() (string, error) {
	selector := "[data-testid=FileViewerHeaderFileName]"

	for {
		res, err := j.page.Evaluate(
			"() => document.querySelector('" + selector + "')",
		)
		if err != nil {
			return "", err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	res, err := j.page.Evaluate(
		"document.querySelector('" + selector + "').childNodes[0].textContent.split('.').slice(0, -1).join('.')",
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", res), nil
}

func (j *Jottacloud) EvaluateFileExt() (string, error) {
	selector := "[data-testid=FileViewerHeaderFileName]"

	for {
		res, err := j.page.Evaluate(
			"() => document.querySelector('" + selector + "')",
		)
		if err != nil {
			return "", err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	res, err := j.page.Evaluate(
		"document.querySelector('" + selector + "').childNodes[0].textContent.split('.').at(-1)",
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(".%v", res), nil
}

func (j *Jottacloud) Download(tempDir, finalDir, filename string, setProgress func(p int8)) error {
	for {
		res, err := j.page.Evaluate(
			"() => document.querySelector('[data-testid=FileViewerHeaderFileName]')",
		)
		if err != nil {
			return err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	fp := filepath.Join(finalDir, filename)
	fileExists, err := appUtils.FileExists(fp)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	href, err := j.page.Evaluate("document.querySelector(\"a[download]\").href")
	if err != nil {
		return err
	}
	downloadUrl, ok := href.(string)
	if !ok {
		return fmt.Errorf("Jottacloud: Couldn't get download url")
	}

	err = appUtils.DownloadFile(downloadUrl, tempDir, fp, setProgress)
	if err != nil {
		return err
	}

	return nil
}
