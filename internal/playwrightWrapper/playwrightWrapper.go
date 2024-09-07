package playwrightWrapper

import (
	"fmt"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/store"
)

type PwContainer struct {
	Playwright     *playwright.Playwright
	Browser        playwright.Browser
	BrowserContext playwright.BrowserContext
}

func WithBrowserType(opts ...string) string {
	o := "chromium"

	for _, opt := range opts {
		o = opt
	}

	return o
}

func WithHeadless(opts ...bool) bool {
	appCfgInt, err := store.GetStore().Get("app-config")
	if err != nil {
		panic(err)
	}
	appConfig := appCfgInt.(*configManager.Config)

	playwrightDebug := appConfig.Dev.PlaywrightDebug

	// fmt.Printf("Playwright dbg: %v\n", playwrightDebug)

	o := !playwrightDebug

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
) (*PwContainer, error) {
	HandleInterrupts := true

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
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
		return nil, fmt.Errorf("Incorrect browser type")
	}

	browser, err := bw.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless:      &headless,
			Timeout:       &timeout,
			HandleSIGHUP:  &HandleInterrupts,
			HandleSIGINT:  &HandleInterrupts,
			HandleSIGTERM: &HandleInterrupts,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Couldn't start a new browser: %v", err)
	}

	ctx, err := browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("Couldn't start a new browser context: %v", err)
	}

	return &PwContainer{
		Playwright:     pw,
		Browser:        browser,
		BrowserContext: ctx,
	}, nil
}

func (pwc *PwContainer) Close() error {
	err := pwc.Browser.Close()
	if err != nil {
		return fmt.Errorf("playwright browser error: %v", err)
	}

	err = pwc.Playwright.Stop()
	if err != nil {
		return fmt.Errorf("playwright driver error: %v", err)
	}

	return nil
}
