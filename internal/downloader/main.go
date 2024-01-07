package downloader

import (
	"fmt"
	"os"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"

	"github.com/playwright-community/playwright-go"
)

const (
	DOUJINSTYLE_ALBUM_URL       = "https://doujinstyle.com/?p=page&type=1&id="
	DEFAULT_DOUJINSTYLE_ERR     = "Doujinstyle.com: 'Insufficient information to display content.'"
	DEFAULT_DOWNLOAD_ERR        = "Not an handled download url, album url: "
	DEFAULT_PAGE_NOT_LOADED_ERR = "The download page did not load in a reasonable amount of time."
)

func createDownloadFolder() error {
	appConfig, err := configManager.NewConfig()
	if err != nil {
		return err
	}
	DOWNLOAD_ROOT := appConfig.Download.Directory

	if _, err := os.Stat(DOWNLOAD_ROOT); os.IsNotExist(err) {
		err = os.MkdirAll(DOWNLOAD_ROOT, 0755)
		if err != nil {
			fmt.Println("Error creating download folder:", err)
			return err
		}
	}

	return nil
}

func handleDownloadPage(albumName string, dlPage playwright.Page) error {
	pageUrl := dlPage.URL()

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

func checkDMCA(p *playwright.Page) (bool, error) {
	locator := (*p).Locator("h3")
	htmlElements, err := locator.All()
	if err != nil {
		return false, err
	}

	text, err := htmlElements[0].InnerHTML()
	if err != nil {
		return false, err
	}

	if text == "Insufficient information to display content." {
		return true, nil
	}

	return false, nil
}

func Download(albumID string, bw *playwright.Browser) error {
	err := createDownloadFolder()
	if err != nil {
		return err
	}

	ctx, err := (*bw).NewContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	_, err = page.Goto(DOUJINSTYLE_ALBUM_URL + albumID)
	if err != nil {
		return err
	}

	err = page.WaitForLoadState()
	if err != nil {
		return fmt.Errorf("Page took too long to load")
	}

	isDMCA, err := checkDMCA(&page)
	if err != nil {
		return err
	}
	if isDMCA {
		return fmt.Errorf(DEFAULT_DOUJINSTYLE_ERR)
	}

	albumName, err := CraftFilename(page)
	if err != nil {
		return err
	}
	// fmt.Printf("Filename: %s\n", albumName)

	dlPage, err := ctx.ExpectPage(func() error {
		_, err := page.Evaluate("document.querySelector('#downloadForm').click()")
		return err
	})
	if err != nil {
		return err
	}

	err = dlPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	if err != nil {
		runBeforeUnloadOpt := true

		pageCloseOptions := playwright.PageCloseOptions{
			RunBeforeUnload: &runBeforeUnloadOpt,
		}

		dlPage.Close(pageCloseOptions)

		return fmt.Errorf(DEFAULT_PAGE_NOT_LOADED_ERR)
	}

	err = handleDownloadPage(albumName, dlPage)

	return err
}
