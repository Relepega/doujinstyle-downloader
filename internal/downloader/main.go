package downloader

import (
	"fmt"
	"net/url"
	"os"

	"github.com/playwright-community/playwright-go"
)

const DOWNLOAD_ROOT = "./Downloads"

func Download(albumID string, ctx *playwright.BrowserContext) error {
	if _, err := os.Stat(DOWNLOAD_ROOT); os.IsNotExist(err) {
		err = os.MkdirAll(DOWNLOAD_ROOT, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	var (
		urlParse      *url.URL
		page          playwright.Page
		dlPage        playwright.Page
		albumName     string
		val           bool
		pageNotExists interface{}
	)

	page, err := (*ctx).NewPage()
	if err != nil {
		return err
	}
	page.Goto("https://doujinstyle.com/?p=page&type=1&id=" + albumID)

	err = page.WaitForLoadState()
	if err != nil {
		return err
	}

	pageNotExists, err = page.Evaluate(
		"document.querySelectorAll('h3')[0].innerText == 'Insufficient information to display content.'",
	)
	if err != nil {
		return err
	}

	val, _ = pageNotExists.(bool)
	if val {
		err = fmt.Errorf("Doujinstyle.com: 'Insufficient information to display content.'")
	}

	albumName, err = CraftFilename(page)
	if err != nil {
		return err
	}
	// fmt.Printf("Filename: %s\n", albumName)

	dlPage, err = (*ctx).ExpectPage(func() error {
		page.Evaluate("document.querySelector('#downloadForm').click()")
		return nil
	})
	if err != nil {
		return err
	}

	err = dlPage.WaitForLoadState()
	if err != nil {
		return err
	}

	// dlPage.On("popup", func(p playwright.Page) {
	// 	p.Close()
	// })

	urlParse, err = url.Parse(dlPage.URL())
	if err != nil {
		return err
	}

	switch urlParse.Hostname() {
	case "www.mediafire.com":
		{
			err = Mediafire(albumName, &dlPage)
		}
	case "mega.nz":
		{
			err = Mega(albumName, &dlPage)
		}
	default:
		{
			err = fmt.Errorf("Not an handled download url, album url: " + dlPage.URL())
		}
	}

	dlPage.Close()
	page.Close()

	return err
}
