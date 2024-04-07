package hosts

import (
	"fmt"
	"os"
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
}

func isFolder(url string) bool {
	if strings.Contains(url, "/folder/") {
		return true
	}

	return false
}

func getFolderKey(url string) string {
	urlElems := strings.Split(url, "/")

	lastUrlElem := len(urlElems) - 1

	folderkey := urlElems[lastUrlElem-1]

	if urlElems[lastUrlElem] == "" {
		folderkey = urlElems[lastUrlElem-2]
	}

	return folderkey
}

func fetchFolderContent(folderKey string, dir string) ([]*fileData, error) {
	fd := []*fileData{}

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

		fd = append(fd, &fileData{
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

		newFd, err := fetchFolderContent(folder.FolderKey, newDir)
		if err != nil {
			return nil, err
		}

		fd = append(fd, newFd[:]...)
	}

	return fd, nil

}

func Mediafire(albumName string, dlPage playwright.Page, progress *int8) error {
	defer dlPage.Close()

	if !isFolder(dlPage.URL()) {
		err := downloadSingleFile(albumName, dlPage, progress, "")
		return err
	}

	folderKey := getFolderKey(dlPage.URL())

	files, err := fetchFolderContent(folderKey, albumName)
	if err != nil {
		return err
	}

	downloadedFiles := 0
	totalFiles := len(files)

	*progress = 0

	appConfig, err := configManager.NewConfig()
	if err != nil {
		return err
	}
	DOWNLOAD_ROOT := appConfig.Download.Directory

	var dummyProg int8

	for _, f := range files {
		p, err := dlPage.Context().NewPage()
		if err != nil {
			return err
		}

		_, err = p.Goto(f.Url)
		if err != nil {
			return err
		}

		dlPath := filepath.Join(DOWNLOAD_ROOT, f.Directory)
		folderExists, _ := appUtils.DirectoryExists(dlPath)
		if !folderExists {
			os.MkdirAll(dlPath, 0755)
		}

		ok, _ := appUtils.FileExists(filepath.Join(dlPath, f.Filename))
		if !ok {
			downloadSingleFile(f.Filename, p, &dummyProg, f.Directory)
		}

		downloadedFiles++
		*progress = int8((float64(downloadedFiles) / float64(totalFiles)) * 100)
	}

	return err
}

func downloadSingleFile(filename string, dlPage playwright.Page, progress *int8, directory string) error {
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

	fp := filepath.Join(DOWNLOAD_ROOT, directory, filename+extension)
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
