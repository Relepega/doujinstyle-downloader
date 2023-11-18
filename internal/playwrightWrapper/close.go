package playwrightwrapper

import (
	"github.com/playwright-community/playwright-go"
)

func ClosePlaywright(
	pw *playwright.Playwright,
	bw playwright.Browser,
	ctx playwright.BrowserContext,
) error {
	err := ctx.Close()
	if err != nil {
		return err
	}

	err = bw.Close()
	if err != nil {
		return err
	}

	err = pw.Stop()
	if err != nil {
		return err
	}

	return nil
}
