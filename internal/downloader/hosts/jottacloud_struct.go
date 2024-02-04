package hosts

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

type jottacloud struct {
	Host
	progress   *int8
	albumName  *string
	fnSelector string
	page       playwright.Page
}

func (host *jottacloud) evaluateFilename() (string, error) {
	return fmt.Sprintf("jottacloud-%v", time.Now().Unix()), nil
}

func (host *jottacloud) evaluateFileExtension() (string, error) {
	res, err := host.page.Evaluate(
		"document.querySelector('" + host.fnSelector + "').childNodes[0].textContent.split('.')[1]",
	)
	if err != nil {
		return "", err
	}

	extension := fmt.Sprintf(".%v", res)

	return extension, nil
}

func (host *jottacloud) waitForPageLoad() error {
	for {
		res, err := host.page.Evaluate(
			"() => document.querySelector('" + host.fnSelector + "')",
		)
		if err != nil {
			return err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	return nil
}

func (host *jottacloud) isFileAvailale() (bool, error) {
	return true, nil
}

func (host *jottacloud) Download(fp string) error {
	href, err := host.page.Evaluate("document.querySelector(\"a[download]\").href")
	if err != nil {
		return err
	}
	downloadUrl, ok := href.(string)
	if !ok {
		return fmt.Errorf("Jottacloud: Couldn't get download url")
	}

	err = appUtils.DownloadFile(fp, downloadUrl, host.progress)
	if err != nil {
		return err
	}

	return nil
}

func (host *jottacloud) Download_dunno_man(albumName string) error {
	defer host.page.Close()

	for {
		res, err := host.page.Evaluate(
			"() => document.querySelector('" + host.fnSelector + "')",
		)
		if err != nil {
			return err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	res, err := host.page.Evaluate(
		"document.querySelector('" + host.fnSelector + "').childNodes[0].textContent.split('.')[1]",
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
	fileExists, err := appUtils.FileExists(fp)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	href, err := host.page.Evaluate("document.querySelector(\"a[download]\").href")
	if err != nil {
		return err
	}
	downloadUrl, ok := href.(string)
	if !ok {
		return fmt.Errorf("Jottacloud: Couldn't get download url")
	}

	err = appUtils.DownloadFile(fp, downloadUrl, host.progress)
	if err != nil {
		return err
	}

	return nil
}
