package task

import (
	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
)

type Task struct {
	Slug         string
	ServiceUrl   string
	HostUrl      string
	TempDir      string
	DownloadsDir string
	Filename     string
	Progress     int8
}

func NewTask() *Task {
	return &Task{
		TempDir: appUtils.GetAppTempDir(),
	}
}

func NewTaskFromServiceURL(serviceUrl string) *Task {
	return &Task{
		ServiceUrl: serviceUrl,
		TempDir:    appUtils.GetAppTempDir(),
	}
}

func NewTaskFromSlug(slug string) *Task {
	return &Task{
		Slug:    slug,
		TempDir: appUtils.GetAppTempDir(),
	}
}
