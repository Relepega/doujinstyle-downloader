package hosts

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// this probably is the better solution...
type Host interface {
	Download(filename string) error
	evaluateFileExtension() (string, error)
	evaluateFilename() (string, error)
	isFileAvailable() (bool, error)
	waitForPageLoad() error
}

// ...but for now we do the ugly thing
type HostFunc (func(albumName string, dlPage playwright.Page, progress *int8) error)

const DEFAULT_DOWNLOAD_ERR = "Not an handled download url: "

func Switch(url string) (HostFunc, error) {
	hostname := strings.Split(url, "/")[2]

	switch hostname {
	case "www.mediafire.com":
		return Mediafire, nil
	case "mega.nz":
		return Mega, nil
	case "drive.google.com":
		return GDrive, nil
	case "www.jottacloud.com":
		return Jottacloud, nil
	default:
		return nil, fmt.Errorf(DEFAULT_DOWNLOAD_ERR + url)
	}
}
