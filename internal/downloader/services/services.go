package services

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type Service interface {
	OpenServicePage(ctx *playwright.BrowserContext) (playwright.Page, error)
	CheckDMCA(p playwright.Page) (bool, error)
	EvaluateFilename(p playwright.Page) (string, error)
	OpenDownloadPage(servicePage playwright.Page) (playwright.Page, error)
}

const SERVICE_ERROR_404 = "Error 404, page not found"

func NewService(service string, mediaID string) (Service, error) {
	switch service {
	case "doujinstyle":
		return newDoujinstyle(mediaID), nil

	case "sukidesuost":
		return newSukidesuost(mediaID), nil

	default:
		return nil, fmt.Errorf("unknown service")
	}
}

func NewServiceFromURL(pageURL string) (Service, error) {
	urlObject, err := url.Parse(pageURL)
	if err != nil {
		log.Fatal(err)
	}

	hostname := strings.TrimPrefix(urlObject.Hostname(), "www.")

	switch hostname {
	case "doujinstyle.com":
		return newDoujinstyle(pageURL), nil

	case "sukidesuost.info":
		return newSukidesuost(pageURL), nil

	default:
		return nil, fmt.Errorf("Unknown service")
	}
}
