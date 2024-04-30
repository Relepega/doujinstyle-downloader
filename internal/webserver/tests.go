package webserver

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/relepega/doujinstyle-downloader/internal/webserver/SSEEvents"
)

func (ws *webserver) handleHello(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "index.html")
	fmt.Fprintf(w, "Hello, World!")
}

func (ws *webserver) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	ws.msgChan <- SSEEvents.NewSSEMessage(fmt.Sprintf("hello, %d", time.Now().Unix()))
}

func (ws *webserver) handleSquare(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("number")

	n, err := strconv.Atoi(m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err.Error())
		return
	}

	err = ws.templates.ExecuteWithWriter(w, "square", struct{ N int }{N: n})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
	}
}
