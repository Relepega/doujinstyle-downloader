package services

import "github.com/playwright-community/playwright-go"

type ServiceNumber int

const (
	Doujinstyle ServiceNumber = iota
)

type Service interface {
	checkDMCA(p *playwright.Page) (bool, error)
	evaluateFilename(p playwright.Page) (string, error)
	Process() error
}

func NewService(
	service ServiceNumber,
	urlSlug string,
	bw *playwright.Browser,
	progress *int8,
) Service {
	switch service {
	case 0:
		return &doujinstyle{
			albumID:  urlSlug,
			bw:       bw,
			progress: progress,
		}
	default:
		return nil
	}
}
