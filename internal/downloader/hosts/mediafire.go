package hosts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

type fileData struct {
	Directory string
	Filename  string
	Url       string
	IsFolder  bool
}

func isFolder(url string) bool {
	if strings.Contains(url, "/folder/") {
		return true
	}

	return false
}

func getFolderFiles() []fileData {
	files := make([]fileData, 0)
	return files
}

func Mediafire(albumName string, dlPage playwright.Page, progress *int8) error {
	defer dlPage.Close()

	var err error

	if isFolder(dlPage.URL()) {
		// err = fmt.Errorf("Mediafire: Folder download is not supported yet.")
		// files := make([]int, 0)
	} else {
		err = file(albumName, dlPage, progress)
	}

	return err
}

func file(albumName string, dlPage playwright.Page, progress *int8) error {
	for {
		res, err := dlPage.Evaluate(
			"() => document.querySelector(\".DownloadStatus.DownloadStatus--uploading\")",
		)
		if err != nil {
			return err
		}

		if res == nil {
			break
		}

		time.Sleep(time.Second * 5)
	}

	var extension string

	ext, _ := dlPage.Evaluate("document.querySelector('.filetype').innerText")
	if ext == nil {
		ext, _ = dlPage.Evaluate(`() => {
			let data = document.querySelector('.dl-btn-label').title.split('.')
			return data[data.length - 1]
		}`)

		extension = fmt.Sprintf(".%v", ext)
	} else {
		extension = fmt.Sprintf("%v", ext)

		re, err := regexp.Compile(`\.[a-zA-Z0-9]+`)
		if err != nil {
			return err
		}
		extension = strings.ToLower(re.FindString(extension))
	}

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

	href, err := dlPage.Evaluate("document.querySelector('#downloadButton').href")
	if err != nil {
		return err
	}
	downloadUrl, ok := href.(string)
	if !ok {
		return fmt.Errorf("Mediafire: Couldn't get download url")
	}

	err = appUtils.DownloadFile(fp, downloadUrl, progress)
	if err != nil {
		return err
	}

	return nil
}
