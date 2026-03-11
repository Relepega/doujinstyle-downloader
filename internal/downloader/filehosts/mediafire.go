package filehosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
)

type Mediafire struct {
	dsdl.Filehost

	page playwright.Page
}

type mediafire_file_data struct {
	Directory string
	Filename  string
	Url       string
}

func NewMediafire(p playwright.Page) dsdl.FilehostImpl {
	return &Mediafire{
		page: p,
	}
}

func (m *Mediafire) SetPage(p playwright.Page) {
	m.page = p
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
	// return m.page.Evaluate(`(() => {
	//        let title = document.querySelector('.dl-btn-label').title
	//        let innerText = document.querySelector('.dl-btn-label').innerText
	//
	//        let start   = innerText.split('').length
	//
	//        return title.split('').slice(start+1).join('')
	//    })() `)

	// bypass if folder, to not error out the whole download
	if m.isFolder() {
		return "", nil
	}

	innerText, err := m.EvaluateFileName()
	if err != nil {
		return "", err
	}

	title_iface, err := m.page.Evaluate("document.querySelector('.dl-btn-label').title")
	if err != nil {
		return "", err
	}

	title, ok := title_iface.(string)
	if !ok {
		return "", fmt.Errorf("Cannot convert data into string")
	}

	ext := title[len(innerText)+1:]

	return ext, nil
}

func (m *Mediafire) Download(tempDir, finalDir, filename string, setProgress func(p int8),
) error {
	if !m.isFolder() {
		err := m.downloadSingleFile(tempDir, finalDir, filename, setProgress)
		return err
	}

	key := m.getFolderKey()

	files, err := m.fetchFolderContent(key, tempDir, filepath.Join(finalDir, filename))
	if err != nil {
		return err
	}

	totalFiles := len(files)
	downloadedFiles := 0

	setProgress(0)

	for _, f := range files {
		_, err = m.page.Goto(f.Url, playwright.PageGotoOptions{
			// Timeout:   new(float64), // default for 'new' is 0.0, which disables timeout
			WaitUntil: playwright.WaitUntilStateCommit,
		})
		if err != nil && strings.Contains(err.Error(), "Timeout") {
			return err
		}

		folderExists := appUtils.DirectoryExists(f.Directory)
		if !folderExists {
			os.MkdirAll(f.Directory, 0o755)
		}

		abs_filename := filepath.Join(f.Directory, f.Filename)

		ok, _ := appUtils.FileExists(abs_filename)
		if !ok {
			err := m.downloadSingleFile(tempDir, f.Directory, f.Filename, func(p int8) {})
			if err != nil {
				return err
			}
		}

		downloadedFiles++
		setProgress(int8((float64(downloadedFiles) / float64(totalFiles)) * 100))
	}

	return err
}

/*

filehost-specific functions

*/

func (m *Mediafire) isFolder() bool {
	return strings.Contains(m.page.URL(), "/folder/")
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

// the finalDir is being treated as the root dir for the folder
func (m *Mediafire) fetchFolderContent(
	folderKey,
	tempFilepath,
	baseDir string,
) ([]*mediafire_file_data, error) {
	fd := []*mediafire_file_data{}

	// parse folders json
	url := fmt.Sprintf(
		"https://www.mediafire.com/api/1.5/folder/get_content.php?content_type=files&version=1.5&folder_key=%s&response_format=json",
		folderKey,
	)

	var apiData MediafireFolderInfoResponse
	err := appUtils.ParseJson(url, &apiData)
	if err != nil {
		return nil, err
	}
	if apiData.Response.Result != "Success" {
		return nil, fmt.Errorf("Mediafire API: Couldn't get folder content")
	}

	for _, f := range apiData.Response.FolderContent.Files {
		if f.PasswordProtected != "no" {
			continue
		}

		if f.Permissions.Read != "1" {
			continue
		}

		fd = append(fd, &mediafire_file_data{
			Directory: baseDir,
			Filename:  f.Filename,
			Url:       f.Links.NormalDownload,
		})

	}

	for _, folder := range apiData.Response.FolderContent.Folders {
		if folder.Permissions.Read != "1" || folder.FileCount == "0" {
			continue
		}

		nestedFinalFP := filepath.Join(baseDir, folder.Name)

		newFd, err := m.fetchFolderContent(folder.FolderKey, tempFilepath, nestedFinalFP)
		if err != nil {
			return nil, err
		}

		fd = append(fd, newFd[:]...)
	}

	return fd, nil
}

func (m *Mediafire) downloadSingleFile(
	tempDir, finalDir, filename string,
	setProgress func(p int8),
) error {
	// file is still in upload status?
	for {
		res, err := m.page.Evaluate(
			`() => document.querySelector(".DownloadStatus.DownloadStatus--uploading")`,
		)
		if err != nil {
			return err
		}

		if res == nil {
			break
		}

		time.Sleep(time.Second * 5)
	}

	finalFilepath := filepath.Join(finalDir, filename)

	fileExists, err := appUtils.FileExists(finalFilepath)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	var downloadUrl string

	retryThreshold := 20

	for {
		if retryThreshold <= 0 {
			return fmt.Errorf(
				"Mediafire.downloadSingleFile: Threshold exceeded: could not fetch the download url in time",
			)
		}

		href, err := m.page.Evaluate(
			// `atob(document.querySelector('#downloadButton').getAttribute("data-scrambled-url"))`,
			`document.querySelector('#downloadButton').href`,
		)
		if err != nil {
			retryThreshold--

			time.Sleep(1 * time.Second)

			continue
		}
		url, ok := href.(string)
		if !ok {
			return fmt.Errorf("Mediafire: Couldn't get download url")
		}

		downloadUrl = url

		break
	}

	err = appUtils.DownloadFile(
		downloadUrl,
		tempDir,
		finalFilepath,
		setProgress,
	)
	if err != nil {
		return err
	}

	return nil
}
