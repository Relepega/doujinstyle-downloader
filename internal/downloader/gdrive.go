package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func GDrive(albumName string, dlPage playwright.Page) error {
	defer dlPage.Close()

	pageUrl := dlPage.URL()

	newPage, err := dlPage.Context().NewPage()

	_, err = newPage.Goto(
		"https://drive.google.com/u/0/uc?id=" + strings.Split(pageUrl, "/")[5] + "&export=download",
	)
	if err != nil {
		return err
	}

	err = newPage.WaitForLoadState()
	if err != nil {
		return err
	}

	res, err := newPage.Evaluate(
		"document.querySelector('a').innerText.split('.').toReversed()[0]",
	)
	if err != nil {
		return err
	}

	extension := fmt.Sprintf(".%v", res)

	fp := filepath.Join(DOWNLOAD_ROOT, albumName+extension)
	_, err = os.Stat(fp)
	if err == nil {
		return nil
	}

	downloadHandler, err := newPage.ExpectDownload(func() error {
		_, err := newPage.Evaluate("document.querySelector('#uc-download-link').click()")
		return err
	})
	if err != nil {
		return err
	}

	_ = newPage.Close()

	time.Sleep(time.Second)

	err = downloadHandler.SaveAs(fp)
	if err != nil {
		return fmt.Errorf("%v\n--------------\n%v", err, downloadHandler.Failure())
	}

	return nil
}
