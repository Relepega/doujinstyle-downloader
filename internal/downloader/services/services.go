package services

import "github.com/playwright-community/playwright-go"

type Service interface {
	checkDMCA(p *playwright.Page) (bool, error)
	evaluateFilename(p playwright.Page) (string, error)
	Process() error
}

const SERVICE_ERROR_404 = "Error 404, page not found"

func NewService(
	serviceNumber int,
	urlSlug string,
	bw *playwright.Browser,
	progress *int8,
) Service {
	switch serviceNumber {
	case 0:
		return &doujinstyle{
			albumID:  urlSlug,
			bw:       bw,
			progress: progress,
		}
	case 1:
		return &sukiDesuOst{
			urlSlug:  urlSlug,
			bw:       bw,
			progress: progress,
		}
	default:
		return nil
	}
}
