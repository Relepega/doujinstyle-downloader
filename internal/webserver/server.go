package v2

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/sse"
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

	isHTTPS bool
	sslKey  string
	sslCert string

	httpServer  *http.Server
	connections *sse.Hub

	templates *templates.Templates

	msgChan      chan string
	closeUpdater chan struct{}
	closeStream  chan struct{}

	engine *dsdl.DSDL
}

func NewWebServer(
	address string,
	port uint16,
	isHTTPS bool,
	sslKey string,
	sslCert string,
	dsdl *dsdl.DSDL,
) *Webserver {
	log.Println("Webserver: Initializing webserver")

	server := &http.Server{}

	webServer := &Webserver{
		address:      address,
		port:         port,
		isHTTPS:      isHTTPS,
		sslKey:       sslKey,
		sslCert:      sslCert,
		httpServer:   server,
		connections:  sse.NewHub(),
		msgChan:      make(chan string),
		closeUpdater: make(chan struct{}, 1),
		closeStream:  make(chan struct{}, 1),
		engine:       dsdl,
	}

	return webServer
}

func (ws *Webserver) buildTemplates() *templates.Templates {
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

	t.AddFunction("GetStateStr", states.GetStateStr)

	dir := filepath.Join(".", "views", "templates")
	err = t.ParseGlob(fmt.Sprintf("%s/*.tmpl", dir))
	if err != nil {
		log.Fatalln("Templates parsing error:", err)
	}

	return t
}

func (ws *Webserver) buildRoutes() *http.ServeMux {
	log.Println("Webserver: Building router")

	ws.templates = ws.buildTemplates()

	mux := http.NewServeMux()

	cssDir := http.Dir(filepath.Join(".", "views", "css"))
	jsDir := http.Dir(filepath.Join(".", "views", "js"))

	// resources
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(cssDir)))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(jsDir)))

	// POST   /task { ids: []string }
	mux.HandleFunc(fmt.Sprintf("POST %s/task", APIGroup), ws.handleTaskAdd)
	// PATCH  /task { mode: "single|multiple|failed", ids: []string }
	mux.HandleFunc(fmt.Sprintf("PATCH %s/task", APIGroup), ws.handleTaskUpdate)
	// DELETE /task { mode: "single|multiple|queued|failed|succeeded", ids: []string }
	mux.HandleFunc(fmt.Sprintf("DELETE %s/task", APIGroup), ws.handleTaskRemove)

	mux.HandleFunc("GET /events-stream", ws.handleEventStream)

	// handle hello test endpoint
	mux.HandleFunc("/hello", ws.handleHelloRoute)

	mux.HandleFunc("/", ws.handleIndexRoute)

	// maintenance
	mux.HandleFunc(fmt.Sprintf("POST %s/restart", InternalGroup), ws.handleRestartServer)

	return mux
}

func (ws *Webserver) Start() error {
	log.Println("Webserver: Starting HTTP server")

	mux := ws.buildRoutes()

	netAddr := fmt.Sprintf("%s:%d", ws.address, ws.port)

	ws.httpServer.Addr = netAddr
	ws.httpServer.Handler = mux

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Webserver: Recovered from panic in sseMessageBroker:", r)
			}
		}()

		ws.sseMessageBroker()
	}()

	if ws.isHTTPS {
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
		}

		ws.httpServer.TLSConfig = tlsConfig

		// Start the server in a goroutine
		go func() {
			if err := ws.httpServer.ListenAndServeTLS(
				ws.sslCert,
				ws.sslKey,
			); !errors.Is(
				err,
				http.ErrServerClosed,
			) {
				log.Fatalf("Webserver: error starting webserver: %v", err)
			}
			log.Println("Webserver: Stopped serving new connections")
		}()

		netAddr = "https://" + netAddr
	} else {
		// Start the server in a goroutine
		go func() {
			if err := ws.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Webserver: error starting webserver: %v", err)
			}
			log.Println("Webserver: Stopped serving new connections")
		}()

		netAddr = "http://" + netAddr
	}

	fmt.Printf("Server is running on %s\n", netAddr)
	log.Printf("Server is running on %s\n", netAddr)

	return nil
}

func (ws *Webserver) Shutdown(ctx context.Context) {
	log.Println("Webserver: Started shutdown procedure")

	ws.connections.Shutdown()

	ws.closeUpdater <- struct{}{}
	close(ws.closeUpdater)

	ws.closeStream <- struct{}{}
	close(ws.closeStream)

	close(ws.msgChan)

	if err := ws.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Webserver: forced shutdown: %v\n", err)
	}

	log.Println("Webserver: exited gracefully")
}

func (ws *Webserver) GetTemplates() templates.Templates {
	return *ws.templates
}
