package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/appUtils"
	pubsub "github.com/relepega/doujinstyle-downloader-reloaded/internal/pubSub"
	eventbroker "github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue/event_broker"
)

type mediafire struct {
	Host

	page playwright.Page

	albumID   string
	albumName string

	dlPath     string
	dlProgress *int8
}

func newMediafire(p playwright.Page, albumID, albumName, downloadPath string, progress *int8) Host {
	return &mediafire{
		page: p,

		albumID:   albumID,
		albumName: albumName,

		dlPath:     downloadPath,
		dlProgress: progress,
	}
}

type mediafire_file_data struct {
	Directory string
	Filename  string
	Url       string
}

func (m *mediafire) isFolder(url string) bool {
	if strings.Contains(url, "/folder/") {
		return true
	}

	return false
}

func (m *mediafire) getFolderKey(url string) string {
	urlElems := strings.Split(url, "/")

	lastUrlElem := len(urlElems) - 1

	folderkey := urlElems[lastUrlElem-1]

	if urlElems[lastUrlElem] == "" {
		folderkey = urlElems[lastUrlElem-2]
	}

	return folderkey
}

func (m *mediafire) fetchFolderContent(folderKey string, dir string) ([]*mediafire_file_data, error) {
	fd := []*mediafire_file_data{}

	// parse folders json
	url := fmt.Sprintf("https://www.mediafire.com/api/1.5/folder/get_content.php?content_type=folders&version=1.5&folder_key=%s&response_format=json", folderKey)

	var foldersData MediafireFolderContent
	err := appUtils.ParseJson[MediafireFolderContent](url, &foldersData)
	if err != nil {
		return nil, err
	}
	if foldersData.Response.Result != "Success" {
		return nil, fmt.Errorf("Mediafire API: Couldn't get folder content")
	}

	// parse files json
	url = fmt.Sprintf("https://www.mediafire.com/api/1.5/folder/get_content.php?content_type=files&version=1.5&folder_key=%s&response_format=json", folderKey)

	var filesData MediafireFolderContent
	err = appUtils.ParseJson[MediafireFolderContent](url, &filesData)
	if err != nil {
		return nil, err
	}
	if filesData.Response.Result != "Success" {
		return nil, fmt.Errorf("Mediafire API: Couldn't get files data")
	}

	for _, f := range filesData.Response.FolderContent.Files {
		if f.PasswordProtected != "no" {
			continue
		}

		if f.Permissions.Read != "1" {
			continue
		}

		splitFn := strings.Split(f.Filename, ".")

		fd = append(fd, &mediafire_file_data{
			Directory: dir,
			Filename:  strings.Join(splitFn[0:len(splitFn)-1], "."),
			Url:       f.Links.NormalDownload,
		})

	}

	for _, folder := range foldersData.Response.FolderContent.Folders {
		if folder.Permissions.Read != "1" || folder.FileCount == "0" {
			continue
		}

		newDir := filepath.Join(dir, folder.Name)

		newFd, err := m.fetchFolderContent(folder.FolderKey, newDir)
		if err != nil {
			return nil, err
		}

		fd = append(fd, newFd[:]...)
	}

	return fd, nil

}

func (m *mediafire) downloadSingleFile(filename string, dlPage playwright.Page, progress *int8) error {
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

	fp := filepath.Join(m.dlPath, filename+extension)
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

	err = appUtils.DownloadFile(
		fp,
		downloadUrl,
		progress,
		func(p int8) {
			pub, _ := pubsub.GetGlobalPublisher("queue")
			pub.Publish(&pubsub.PublishEvent{
				EvtType: "update-task-progress",
				Data: &eventbroker.UpdateTaskProgress{
					Id:       m.albumID,
					Progress: p,
				},
			})
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *mediafire) Download() error {
	if !m.isFolder(m.page.URL()) {
		err := m.downloadSingleFile(m.albumName, m.page, m.dlProgress)
		return err
	}

	m.dlPath = filepath.Join(m.dlPath, m.albumName)
	err := appUtils.CreateFolder(m.dlPath)
	if err != nil {
		return err
	}

	folderKey := m.getFolderKey(m.page.URL())

	files, err := m.fetchFolderContent(folderKey, m.albumName)
	if err != nil {
		return err
	}

	downloadedFiles := 0
	totalFiles := len(files)

	*m.dlProgress = 0

	var dummyProg int8

	for _, f := range files {
		p, err := m.page.Context().NewPage()
		if err != nil {
			return err
		}

		_, err = p.Goto(f.Url)
		if err != nil {
			return err
		}

		dlPath := filepath.Join(m.dlPath, f.Directory)
		folderExists, _ := appUtils.DirectoryExists(dlPath)
		if !folderExists {
			os.MkdirAll(dlPath, 0755)
		}

		ok, _ := appUtils.FileExists(filepath.Join(dlPath, f.Filename))
		if !ok {
			m.downloadSingleFile(f.Filename, p, &dummyProg)
		}

		downloadedFiles++
		*m.dlProgress = int8((float64(downloadedFiles) / float64(totalFiles)) * 100)
	}

	return err
}
