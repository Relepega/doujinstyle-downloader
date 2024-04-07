package webserver

import (
	"fmt"
	"net/http"
)

var connectedClients = make(map[string]http.ResponseWriter)

func (ws *webserver) handleEventStream(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Client connected")

	WriteDefaultHeaders(w)

	for {
		select {
		case msg := <-ws.msgChan:
			s := msg.String()
			fmt.Fprintf(w, s)
			w.(http.Flusher).Flush()

		case <-r.Context().Done():
			fmt.Println("Client disconnected")
			return

		default:
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
