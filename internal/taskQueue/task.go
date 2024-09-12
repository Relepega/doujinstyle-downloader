package taskQueue

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/hosts"
	"github.com/relepega/doujinstyle-downloader/internal/downloader/services"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	"github.com/relepega/doujinstyle-downloader/internal/store"
)

type Task struct {
	ctx context.Context

	AlbumID     string
	DisplayName string
	Service     string
	Active      bool
	Done        bool
	Error       error

	IsChecking bool

	DownloadProgress int8
}

func NewTask(AlbumID string, service string) *Task {
	return &Task{
		AlbumID:     AlbumID,
		DisplayName: AlbumID,
		Service:     service,

		Active: false,
		Done:   false,

		Error: nil,

		IsChecking: false,

		DownloadProgress: -1,
	}
}

func (t *Task) Activate() {
	t.Active = true
}

func (t *Task) Deactivate() {
	t.Active = false
}

func (t *Task) MarkAsDone(e error) {
	t.Active = false
	t.Done = true
	t.Error = e
}

func (t *Task) Reset() {
	if t.Active {
		return
	}

	t.Active = false
	t.Done = false
	t.Error = nil
}

func (t *Task) Run(q *Queue, pwc *playwrightWrapper.PwContainer) error {
	q.publishUIUpdate("activate-task", t)

	runBeforeUnloadOpt := true
	pageCloseOpts := playwright.PageCloseOptions{
		RunBeforeUnload: &runBeforeUnloadOpt,
	}

	appCfgInt, err := store.GetStore().Get("app-config")
	if err != nil {
		panic(err)
	}
	appConfig := appCfgInt.(*configManager.Config)

	ctx, err := pwc.Browser.NewContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	service, err := services.NewService(t.Service, t.AlbumID)
	if err != nil {
		return err
	}

	servicePage, err := service.OpenServicePage(&ctx)
	if err != nil {
		return err
	}
	defer servicePage.Close()

	isDMCA, err := service.CheckDMCA(servicePage)
	if err != nil {
		return err
	}

	if isDMCA {
		return fmt.Errorf("%s: %s", t.Service, services.SERVICE_ERROR_404)
	}

	mediaName, err := service.EvaluateFilename(servicePage)
	if err != nil {
		return err
	}

	t.DisplayName = mediaName
	q.publishUIUpdate("update-task-content", t)

	t.IsChecking = true
	alreadyInList := q.isTaskInList(t)
	t.IsChecking = false

	if alreadyInList {
		return fmt.Errorf("Task already done or already in download")
	}

	downloadPage, err := service.OpenDownloadPage(servicePage)
	if err != nil {
		return err
	}
	defer downloadPage.Close(pageCloseOpts)

	_ = servicePage.Close(pageCloseOpts)

	hostFactory, err := hosts.NewHost(downloadPage.URL())
	if err != nil {
		return err
	}

	host := hostFactory(
		downloadPage,
		t.AlbumID,
		mediaName,
		appConfig.Download.Directory,
		&t.DownloadProgress,
	)
	err = host.Download()
	if err != nil {
		return err
	}

	return nil
}
