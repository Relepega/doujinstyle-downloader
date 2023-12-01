package downloader

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

const (
	DOWNLOAD_ROOT           = "./Downloads"
	DOUJINSTYLE_ALBUM_URL   = "https://doujinstyle.com/?p=page&type=1&id="
	DEFAULT_DOUJINSTYLE_ERR = "Doujinstyle.com: 'Insufficient information to display content.'"
	DEFAULT_DOWNLOAD_ERR    = "Not an handled download url, album url: "
)

func createDownloadFolder() error {
	if _, err := os.Stat(DOWNLOAD_ROOT); os.IsNotExist(err) {
		err = os.MkdirAll(DOWNLOAD_ROOT, 0755)
		if err != nil {
			fmt.Println("Error creating download folder:", err)
			return err
		}
	}

	return nil
}

func handleDownloadPage(albumName string, dlPage *playwright.Page) error {
	pageUrl := (*dlPage).URL()

	dlPage_hostname := strings.Split(pageUrl, "/")[2]

	switch dlPage_hostname {
	case "www.mediafire.com":
		return Mediafire(albumName, dlPage)
	case "mega.nz":
		return Mega(albumName, dlPage)
	case "drive.google.com":
		return GDrive(albumName, dlPage)
	case "www.jottacloud.com":
		return Jottacloud(albumName, dlPage)
	default:
		return fmt.Errorf(DEFAULT_DOWNLOAD_ERR + pageUrl)
	}
}

func Download(albumID string, ctx *playwright.BrowserContext) error {
	var (
		page          playwright.Page
		dlPage        playwright.Page
		albumName     string
		val           bool
		pageNotExists interface{}
	)

	err := createDownloadFolder()
	if err != nil {
		return err
	}

	page, err = (*ctx).NewPage()
	if err != nil {
		_ = page.Close()
		return err
	}

	page.Goto(DOUJINSTYLE_ALBUM_URL + albumID)

	err = page.WaitForLoadState()
	if err != nil {
		_ = page.Close()
		return err
	}
	time.Sleep(time.Second)

	pageNotExists, err = page.Evaluate(
		"document.querySelectorAll('h3')[0].innerText == 'Insufficient information to display content.'",
	)
	if err != nil {
		_ = page.Close()
		return err
	}

	val, _ = pageNotExists.(bool)
	if val {
		err = fmt.Errorf(DEFAULT_DOUJINSTYLE_ERR)
	}
	if err != nil {
		_ = page.Close()
		return err
	}

	albumName, err = CraftFilename(page)
	if err != nil {
		_ = page.Close()
		return err
	}
	// fmt.Printf("Filename: %s\n", albumName)

	dlPage, err = (*ctx).ExpectPage(func() error {
		_, err := page.Evaluate("document.querySelector('#downloadForm').click()")
		return err
	})
	if err != nil {
		_ = dlPage.Close()
		_ = page.Close()
		return err
	}

	err = dlPage.WaitForLoadState()
	if err != nil {
		_ = page.Close()
		_ = dlPage.Close()
		return err
	}
	time.Sleep(time.Second)

	err = handleDownloadPage(albumName, &dlPage)

	_ = page.Close()
	_ = dlPage.Close()

	return err
}
