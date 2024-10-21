package v2

import (
	"log"
	"net/http"
)

func (ws *Webserver) handleEventStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := ws.connections.AddClient(w)
	log.Println("New client connected")

	for {
		select {
		case msg := <-ws.msgChan:
			s := msg.String()

			// fmt.Fprintf(w, s)
			// w.(http.Flusher).Flush()
			ws.connections.Broadcast(s)

		case <-r.Context().Done():
			ws.connections.Removeclient(client)
			log.Println("Client disconnected")
			return

		default:
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
