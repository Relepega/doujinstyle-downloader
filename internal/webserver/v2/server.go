package v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	tlp "github.com/relepega/doujinstyle-downloader/internal/topLevelProxy"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/SSEEvents"
	ssehub "github.com/relepega/doujinstyle-downloader/internal/webserver/SSEHub"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/templates"
)

const (
	APIGroup      = "/api"
	TaskGroup     = APIGroup + "/task"
	InternalGroup = APIGroup + "/internal"
)

type Webserver struct {
	address string
	port    uint16

	httpServer  *http.Server
	connections ssehub.ConnectionsHub

	templates *templates.Templates

	msgChan chan *SSEEvents.SSEMessage

	ctx context.Context
}

func NewWebServer(address string, port uint16, ctx context.Context) *Webserver {
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
		log.Fatalln("Templates parsing error:", err)
	}

	webServer := &Webserver{
		address: address,
		port:    port,

		httpServer: server,

		templates: t,

		ctx: ctx,
	}

	webServer.msgChan = make(chan *SSEEvents.SSEMessage)

	return webServer
}

func (ws *Webserver) buildRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	cssDir := http.Dir(filepath.Join(".", "views", "css"))
	jsDir := http.Dir(filepath.Join(".", "views", "js"))

	// resources
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(cssDir)))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(jsDir)))

	// POST   /tasks/add { ids: []string }
	mux.HandleFunc("POST /tasks/add", ws.handleTaskAdd)
	// PATCH  /tasks/update { mode: "single|multiple|failed", ids: []string }
	mux.HandleFunc("POST /tasks/update", ws.handleTaskUpdate)
	// DELETE /tasks/delete { mode: "single|multiple|queued|failed|succeeded", ids: []string }
	mux.HandleFunc("DELETE /tasks/delete", ws.handleTaskRemove)

	mux.HandleFunc("/", ws.handleIndexRoute)

	// maintenance
	mux.HandleFunc(fmt.Sprintf("POST %s/restart", InternalGroup), ws.handleRestartServer)

	return mux
}

func (ws *Webserver) Start(ctx context.Context) error {
	defer close(ws.msgChan)

	mux := ws.buildRoutes()

	// go ws.SSEMsgBroker()

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

func (ws *Webserver) Context() *tlp.TLProxy {
	v, _ := ws.ctx.Value("tlp").(*tlp.TLProxy)
	return v
}
