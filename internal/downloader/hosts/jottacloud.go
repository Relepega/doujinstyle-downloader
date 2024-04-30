package hosts

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	tq_eventbroker "github.com/relepega/doujinstyle-downloader/internal/taskQueue/tq_event_broker"
)

type jottacloud struct {
	Host

	page playwright.Page

	albumID   string
	albumName string

	dlPath     string
	dlProgress *int8
}

func newJottacloud(
	p playwright.Page,
	albumID, albumName, downloadPath string,
	progress *int8,
) Host {
	return &jottacloud{
		page: p,

		albumID:   albumID,
		albumName: albumName,

		dlPath:     downloadPath,
		dlProgress: progress,
	}
}

func (j *jottacloud) Download() error {
	fnSelector := "[data-testid=FileViewerHeaderFileName]"

	for {
		res, err := j.page.Evaluate(
			"() => document.querySelector('" + fnSelector + "')",
		)
		if err != nil {
			return err
		}

		if res != nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	res, err := j.page.Evaluate(
		"document.querySelector('" + fnSelector + "').childNodes[0].textContent.split('.')[1]",
	)
	if err != nil {
		return err
	}

	extension := fmt.Sprintf(".%v", res)

	fp := filepath.Join(j.dlPath, j.albumName+extension)
	fileExists, err := appUtils.FileExists(fp)
	if err != nil {
		return err
	}
	if fileExists {
		return nil
	}

	href, err := j.page.Evaluate("document.querySelector(\"a[download]\").href")
	if err != nil {
		return err
	}
	downloadUrl, ok := href.(string)
	if !ok {
		return fmt.Errorf("Jottacloud: Couldn't get download url")
	}

	err = appUtils.DownloadFile(
		fp,
		downloadUrl,
		j.dlProgress,
		func(p int8) {
			pub, _ := pubsub.GetGlobalPublisher("queue")
			pub.Publish(&pubsub.PublishEvent{
				EvtType: "update-task-progress",
				Data: &tq_eventbroker.UpdateTaskProgress{
					Id:       j.albumID,
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
