package playwrightwrapper

import (
	"github.com/playwright-community/playwright-go"
)

func ClosePlaywright(
	pw *playwright.Playwright,
	bw playwright.Browser,
) error {
	contexts := bw.Contexts()

	for i := 0; i < len(contexts); i++ {
		err := contexts[i].Close()
		if err != nil {
			return err
		}
	}

	err := bw.Close()
	if err != nil {
		return err
	}

	err = pw.Stop()
	if err != nil {
		return err
	}

	return nil
}
