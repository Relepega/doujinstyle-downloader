package webserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/SSEEvents"
	ssehub "github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/SSEHub"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/templates"
)

const (
	APIGroup  = "/api"
	TaskGroup = APIGroup + "/task"
)

type webserver struct {
	address string
	port    uint16

	httpServer  *http.Server
	connections ssehub.ConnectionsHub

	templates *templates.Templates

	msgChan chan *SSEEvents.SSEMessage

	q *taskQueue.Queue
}

func NewWebServer(address string, port uint16, queue *taskQueue.Queue) *webserver {
	server := &http.Server{}

	t, err := templates.NewTemplates()
	if err != nil {
		log.Fatalln(err)
	}

	t.AddFunction("Square", func(n int) int {
		return n * n
	})

	t.AddFunction("Timestamp", func() string {
		return fmt.Sprintf("%d", time.Now().Unix())
	})

	dir := filepath.Join(".", "views", "templates")
	err = t.ParseGlob(fmt.Sprintf("%s/*.tmpl", dir))
	if err != nil {
		log.Fatalln(err)
	}

	webServer := &webserver{
		address: address,
		port:    port,

		httpServer: server,

		templates: t,

		q: queue,
	}

	webServer.msgChan = make(chan *SSEEvents.SSEMessage)

	return webServer
}

func (ws *webserver) buildRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	cssDir := http.Dir(filepath.Join(".", "views", "css"))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(cssDir)))

	jsDir := http.Dir(filepath.Join(".", "views", "js"))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(jsDir)))

	mux.HandleFunc("/", ws.handleIndexRoute)

	mux.HandleFunc(fmt.Sprintf("POST %s", TaskGroup), ws.handleTaskAdd)
	mux.HandleFunc(fmt.Sprintf("DELETE %s", TaskGroup), ws.handleTaskDelete)
	mux.HandleFunc(fmt.Sprintf("PATCH %s", TaskGroup), ws.handleTaskRetry)

	mux.HandleFunc("GET /events-stream", ws.handleEventStream)

	mux.HandleFunc("GET /hello", ws.handleHello)

	mux.HandleFunc("GET /square", ws.handleSquare)

	mux.HandleFunc("POST /send-message", ws.handleSendMessage)

	return mux
}

func (ws *webserver) Start(ctx context.Context) error {
	defer close(ws.msgChan)

	mux := ws.buildRoutes()

	go ws.SSEMsgBroker()

	netAddr := fmt.Sprintf("%s:%d", ws.address, ws.port)

	ws.httpServer.Addr = netAddr
	ws.httpServer.Handler = mux

	// Start the server in a goroutine
	go func() {
		if err := ws.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	fmt.Printf("Server is running on http://%s\n", netAddr)

	// Wait for either the context to be cancelled or for the server to stop serving new connections
	select {
	case <-ctx.Done():
		// Context was cancelled, start the graceful shutdown
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		if err := ws.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP shutdown error: %v", err)
		}

		log.Println("Graceful webserver shutdown complete.")
		return nil
	}
}
