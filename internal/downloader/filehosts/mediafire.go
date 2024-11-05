package filehosts

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

// \S*mediafire.com\S*

type Mediafire struct {
	dsdl.Filehost

	page playwright.Page
}

type mediafire_file_data struct {
	Directory string
	Filename  string
	Url       string
}

func NewMediafire() dsdl.FilehostImpl {
	return &Mediafire{}
}

func (m *Mediafire) Page() playwright.Page {
	return m.page
}

func (m *Mediafire) EvaluateFileName() (string, error) {
	fn_intf, err := m.page.Evaluate("document.querySelector('.dl-btn-label').innerText")
	if err != nil {
		return "", err
	}

	fn, ok := fn_intf.(string)
	if !ok {
		return "", fmt.Errorf("Cannot convert data into string")
	}

	return appUtils.CleanString(fn), nil
}

func (m *Mediafire) EvaluateFileExt() (string, error) {
	fn, err := m.EvaluateFileName()
	if err != nil {
		return "", err
	}

	ext_intf, err := m.page.Evaluate("document.querySelector('.dl-btn-label').innerText")
	if err != nil {
		return "", err
	}

	ext_str, ok := ext_intf.(string)
	if !ok {
		return "", fmt.Errorf("Cannot convert data into string")
	}

	ext := ext_str[len(fn)+1:]

	return ext, nil
}

func (m *Mediafire) Download(downloadPath string, progress *int8) error {
	if !m.isFolder() {
		err := m.downloadSingleFile(downloadPath, progress)
		return err
	}

	var dummyProgress int8

	return nil
}

/*

filehost-specific functions

*/

func (m *Mediafire) isFolder() bool {
	if strings.Contains(m.page.URL(), "/folder/") {
		return true
	}

	return false
}

func (m *Mediafire) getFolderKey() string {
	urlElems := strings.Split(m.page.URL(), "/")

	lastUrlElem := len(urlElems) - 1

	folderkey := urlElems[lastUrlElem-1]

	if urlElems[lastUrlElem] == "" {
		folderkey = urlElems[lastUrlElem-2]
	}

	return folderkey
}

func (m *Mediafire) fetchFolderContent(
	folderKey string,
	dir string,
) ([]*mediafire_file_data, error) {
	fd := []*mediafire_file_data{}

	// parse folders json
	url := fmt.Sprintf(
		"https://www.mediafire.com/api/1.5/folder/get_content.php?content_type=folders&version=1.5&folder_key=%s&response_format=json",
		folderKey,
	)

	var foldersData MediafireFolderInfoResponse
	err := appUtils.ParseJson(url, &foldersData)
	if err != nil {
		return nil, err
	}
	if foldersData.Response.Result != "Success" {
		return nil, fmt.Errorf("Mediafire API: Couldn't get folder content")
	}

	// parse files json
	url = fmt.Sprintf(
		"https://www.mediafire.com/api/1.5/folder/get_content.php?content_type=files&version=1.5&folder_key=%s&response_format=json",
		folderKey,
	)

	var filesData MediafireFolderInfoResponse
	err = appUtils.ParseJson(url, &filesData)
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

func (m *Mediafire) downloadSingleFile(fp string, progress *int8) error {
	// file is still in upload status?
	for {
		res, err := m.page.Evaluate(
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

	fileExists, err := appUtils.FileExists(fp)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	href, err := m.page.Evaluate("document.querySelector('#downloadButton').href")
	if err != nil {
		fmt.Println("it's me, a deferred button render!")
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
			// pub, _ := pubsub.GetGlobalPublisher("queue")
			// pub.Publish(&pubsub.PublishEvent{
			// 	EvtType: "update-task-progress",
			// 	Data: &tq_eventbroker.UpdateTaskProgress{
			// 		Id:       m.albumID,
			// 		Progress: p,
			// 	},
			// })
		},
		false,
	)
	if err != nil {
		return err
	}

	return nil
}
