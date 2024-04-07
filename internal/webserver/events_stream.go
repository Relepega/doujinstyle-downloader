package webserver

import (
	"fmt"
	"net/http"
)

func (ws *webserver) handleEventStream(w http.ResponseWriter, r *http.Request) {
	WriteDefaultHeaders(w)

	for {
		select {
		case msg := <-ws.msgChan:
			s := msg.String()

			fmt.Fprintf(w, s)
			w.(http.Flusher).Flush()

		case <-r.Context().Done():
			return

		default:
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
