package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	taskqueue "relepega/doujinstyle-downloader/internal/taskQueue"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/playwright-community/playwright-go"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("./views/*.html")),
	}
}

const SERVICE_URL = "127.0.0.1:42069"

func main() {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("Couldn't install playwright dependencies: %v", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	q := taskqueue.NewQueue(2)
	q.Run(interrupt)

	e := echo.New()

	templates := NewTemplates()
	e.Renderer = templates
	// e.Use(middleware.Logger())

	e.Static("/css", "./views/css")
	e.Static("/js", "./views/js")

	e.GET("/renderTasks", func(c echo.Context) error {
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/removeTask", func(c echo.Context) error {
		albumID := c.QueryParam("id")

		q.RemoveTask(albumID)

		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/redoTask", func(c echo.Context) error {
		albumID := c.QueryParam("id")

		q.ResetTask(albumID)

		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.POST("/addTask", func(c echo.Context) error {
		albumID := c.FormValue("AlbumID")
		t := taskqueue.NewTask(albumID)

		if q.IsInList(t) {
			return c.String(
				http.StatusInternalServerError,
				"AlbumID already processed or in queue.",
			)
		}

		q.AddTask(t)

		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/clearQueue", func(c echo.Context) error {
		q.ClearQueuedTasks()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/clearAllCompleted", func(c echo.Context) error {
		q.ClearAllCompleted()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/clearSuccessfullyCompleted", func(c echo.Context) error {
		q.ClearSuccessfullyCompleted()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/clearFailedCompleted", func(c echo.Context) error {
		q.ClearFailedCompleted()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/retryFailed", func(c echo.Context) error {
		q.ResetFailedTasks()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", q.NewQueueFree())
	})

	e.Logger.Fatal(e.Start(SERVICE_URL))
	<-interrupt
}
