package hosts

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type HostFactory func(p playwright.Page, albumName, downloadPath string, progress *int8) Host

type Host interface {
	Download() error
}

const DEFAULT_DOWNLOAD_ERR = "Not an handled download url"

func NewHost(pageUrl string) (HostFactory, error) {
	urlObject, err := url.Parse(pageUrl)
	if err != nil {
		log.Fatal(err)
	}

	hostname := strings.TrimPrefix(urlObject.Hostname(), "www.")

	fmt.Println("hostname: ", hostname)

	switch hostname {
	case "mediafire.com":
		return newMediafire, nil

	default:
		return nil, fmt.Errorf("%s: %s", DEFAULT_DOWNLOAD_ERR, hostname)
	}
}
