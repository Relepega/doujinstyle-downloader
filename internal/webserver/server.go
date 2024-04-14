package webserver

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/SSEEvents"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/templates"
)

const (
	APIGroup  = "/api"
	TaskGroup = APIGroup + "/task"
)

type webserver struct {
	address string
	port    uint16

	httpClient *http.Client

	templates *templates.Templates

	msgChan chan *SSEEvents.SSEMessage

	q *taskQueue.Queue
}

func NewWebServer(address string, port uint16, queue *taskQueue.Queue) *webserver {
	client := &http.Client{}

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

		httpClient: client,

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

func (ws *webserver) Start() error {
	defer close(ws.msgChan)

	mux := ws.buildRoutes()

	go ws.SSEMsgBroker()

	fmt.Printf("Server is running on http://%s:%d\n", ws.address, ws.port)

	return http.ListenAndServe(fmt.Sprintf("%s:%d", ws.address, ws.port), mux)
}
