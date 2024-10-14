package task

import (
	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
)

type Task struct {
	ServiceUrl   string
	HostUrl      string
	TempDir      string
	DownloadsDir string
	Filename     string
}

func NewTask() (*Task, error) {
	return &Task{
		TempDir: appUtils.GetAppTempDir(),
	}, nil
}

func NewTaskFromServiceURL(serviceUrl string) (*Task, error) {
	return &Task{
		ServiceUrl: serviceUrl,
		TempDir:    appUtils.GetAppTempDir(),
	}, nil
}
