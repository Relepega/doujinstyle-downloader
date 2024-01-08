package webserver

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func StartWebserver() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	appConfig, err := configManager.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}

	q := taskQueue.NewQueue(int(appConfig.Download.ConcurrentJobs))

	go func(interrupt chan os.Signal) {
		q.Run(interrupt)
	}(interrupt)

	e := echo.New()

	templates := NewTemplates()
	e.Renderer = templates

	if appConfig.Dev.ServerLogging {
		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:   true,
			LogURI:      true,
			LogError:    true,
			HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				const defaultLog = "[HTTP REQUEST] uri=%v status=%v"

				if v.Error == nil {
					slog.Info(fmt.Sprintf(defaultLog, v.URI, v.Status))
				} else {
					slog.Error(fmt.Sprintf(defaultLog+" %v", v.URI, v.Status, v.Error.Error()))
				}

				return nil
			},
		}))
	}

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
		t := taskQueue.NewTask(albumID)

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

	serverAddress := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)

	e.Logger.Fatal(e.Start(serverAddress))
	<-interrupt
}
