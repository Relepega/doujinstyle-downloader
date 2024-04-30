package hosts

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	tq_eventbroker "github.com/relepega/doujinstyle-downloader/internal/taskQueue/tq_event_broker"
)

type gdrive struct {
	Host

	page playwright.Page

	albumID   string
	albumName string

	dlPath     string
	dlProgress *int8
}

func newGDrive(p playwright.Page, albumID, albumName, downloadPath string, progress *int8) Host {
	return &gdrive{
		page: p,

		albumID:   albumID,
		albumName: albumName,

		dlPath:     downloadPath,
		dlProgress: progress,
	}
}

func querySelectorVal(p playwright.Page, eval string) (string, error) {
	valInterface, err := p.Evaluate(eval)
	if err != nil {
		return "", err
	}

	val, _ := valInterface.(string)

	return val, nil
}

func craftDirectDownloadLink(p playwright.Page) (string, error) {
	var id string
	var export string
	var confirm string
	var uuid string
	var err error

	id, err = querySelectorVal(p, `document.querySelector('input[name="id"]').value`)
	if err != nil {
		return "", err
	}

	export, err = querySelectorVal(p, `document.querySelector('input[name="export"]').value`)
	if err != nil {
		return "", err
	}

	confirm, err = querySelectorVal(p, `document.querySelector('input[name="confirm"]').value`)
	if err != nil {
		return "", err
	}

	uuid, err = querySelectorVal(p, `document.querySelector('input[name="uuid"]').value`)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintln(
		"https://drive.usercontent.google.com/download?id=" + id + "&export=" + export + "&confirm=" + confirm + "&uuid=" + uuid,
	)

	url = strings.TrimSpace(url)

	return url, nil
}

func (gd *gdrive) Download() error {
	pageUrl := gd.page.URL()

	_, err := gd.page.Goto(
		"https://drive.google.com/u/0/uc?id=" + strings.Split(pageUrl, "/")[5] + "&export=download",
	)
	if err != nil {
		return err
	}

	res, err := gd.page.Evaluate(
		"document.querySelector('a').innerText.split('.').toReversed()[0]",
	)
	if err != nil {
		return err
	}

	extension := fmt.Sprintf(".%v", res)

	fp := filepath.Join(gd.dlPath, gd.albumName+extension)
	fileExists, _ := appUtils.FileExists(fp)
	if fileExists {
		return nil
	}

	err = gd.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	if err != nil {
		return err
	}

	dlUrl, err := craftDirectDownloadLink(gd.page)
	if err != nil {
		return err
	}

	err = appUtils.DownloadFile(
		fp,
		dlUrl,
		gd.dlProgress,
		func(p int8) {
			pub, _ := pubsub.GetGlobalPublisher("queue")
			pub.Publish(&pubsub.PublishEvent{
				EvtType: "update-task-progress",
				Data: &tq_eventbroker.UpdateTaskProgress{
					Id:       gd.albumID,
					Progress: p,
				},
			})
		},
	)
	if err != nil {
		return err
	}

	return nil
}
