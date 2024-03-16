package webserver

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/configManager"
	"github.com/relepega/doujinstyle-downloader/internal/playwrightWrapper"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func serverLoggerHandler(c echo.Context, v middleware.RequestLoggerValues) error {
	const defaultLog = "[HTTP REQUEST] uri=%v status=%v"

	if v.Error == nil {
		slog.Info(fmt.Sprintf(defaultLog, v.URI, v.Status))
	} else {
		slog.Error(fmt.Sprintf(defaultLog+" %v", v.URI, v.Status, v.Error.Error()))
	}

	return nil
}

func StartWebserver() {
	appConfig, err := configManager.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// Setup
	q := taskQueue.NewQueue(int(appConfig.Download.ConcurrentJobs))
	e := echo.New()

	templates := NewTemplates()
	e.Renderer = templates

	if appConfig.Dev.ServerLogging {
		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:     true,
			LogURI:        true,
			LogError:      true,
			HandleError:   true, // forwards error to the global error handler, so it can decide appropriate status code
			LogValuesFunc: serverLoggerHandler,
		}))
	}

	e.Static("/css", "./views/css")
	e.Static("/js", "./views/js")

	apiGroup := e.Group("/api")

	taskGroup := apiGroup.Group("/task")
	queueGroup := apiGroup.Group("/queue")

	taskGroup.GET("/render", func(c echo.Context) error {
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	taskGroup.GET("/remove", func(c echo.Context) error {
		albumID := c.QueryParam("id")

		q.RemoveTask(albumID)

		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	taskGroup.GET("/retry", func(c echo.Context) error {
		albumID := c.QueryParam("id")

		q.ResetTask(albumID)

		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	taskGroup.POST("/add", func(c echo.Context) error {
		albumID := c.FormValue("AlbumID")
		sNumberStr := c.FormValue("ServiceNumber")

		sNumberInt, err := strconv.Atoi(sNumberStr)
		if err != nil {
			return c.String(
				http.StatusInternalServerError,
				err.Error(),
			)
		}

		t := taskQueue.NewTask(albumID, sNumberInt)

		if q.IsInList(t) {
			return c.String(
				http.StatusInternalServerError,
				"AlbumID already processed or in queue.",
			)
		}

		q.AddTask(t)

		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	queueGroup.GET("/clear", func(c echo.Context) error {
		q.ClearQueuedTasks()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	queueGroup.GET("/clearAllCompleted", func(c echo.Context) error {
		q.ClearAllCompleted()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	queueGroup.GET("/clearSuccessfullyCompleted", func(c echo.Context) error {
		q.ClearSuccessfullyCompleted()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	queueGroup.GET("/clearFailedCompleted", func(c echo.Context) error {
		q.ClearFailedCompleted()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	queueGroup.GET("/retryFailed", func(c echo.Context) error {
		q.ResetFailedTasks()
		return c.Render(http.StatusOK, "tasks", q.NewQueueFree())
	})

	e.GET("/updateInterval", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("%f", appConfig.WebUi.UpdateInterval))
	})

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", q.NewQueueFree())
	})

	serverAddress := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)

	// Start queue and server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	playwrightContainer, err := playwrightWrapper.UsePlaywright(
		playwrightWrapper.WithBrowserType(),
		playwrightWrapper.WithHeadless(),
		playwrightWrapper.WithTimeout(),
	)
	if err != nil {
		log.Fatalf("Cannot open playwright browser: %v", err)
	}

	go func(playwrightContainer *playwrightWrapper.PwContainer) {
		q.Run(playwrightContainer)
	}(playwrightContainer)

	go func(serverAddress string) {
		if err := e.Start(serverAddress); err != nil && err != http.ErrServerClosed {
			log.Fatalln("Shutting down the server")
		}
	}(serverAddress)

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	q.Shutdown(playwrightContainer)

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
