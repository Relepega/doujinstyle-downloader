package playwrightwrapper

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

const PLAYWRIGHT_DEBUG = false

func WithBrowserType(opts ...string) string {
	o := "chromium"

	for _, opt := range opts {
		o = opt
	}

	return o
}

func WithHeadless(opts ...bool) bool {
	o := !PLAYWRIGHT_DEBUG

	for _, opt := range opts {
		o = opt
	}

	return o
}

func WithTimeout(opts ...float64) float64 {
	o := 0.0

	for _, opt := range opts {
		o = opt
	}

	return o
}

func UsePlaywright(
	browserType string,
	headless bool,
	timeout float64,
) (*playwright.Playwright, playwright.Browser, playwright.BrowserContext, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start playwright: %v", err)
	}

	var bw playwright.BrowserType

	switch browserType {
	case "chromium":
		bw = pw.Chromium
	case "firefox":
		bw = pw.Firefox
	case "webkit":
		bw = pw.WebKit
	default:
		return nil, nil, nil, fmt.Errorf("Incorrect browser type")
	}

	browser, err := bw.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: &headless,
			Timeout:  &timeout,
		},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Couldn't start a new browser: %v", err)
	}

	ctx, err := browser.NewContext()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Couldn't start a new browser context: %v", err)
	}

	return pw, browser, ctx, nil
}
