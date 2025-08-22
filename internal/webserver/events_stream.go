package v2

import (
	"fmt"
	"log"
	"net/http"

	"github.com/relepega/doujinstyle-downloader/internal/webserver/sse"
)

func (ws *Webserver) handleEventStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := ws.connections.AddClient(w, r)
	log.Printf("Webserver: EventStream: New client connected (ID: %v)", client.ID())

	for {
		select {
		case <-ws.closeStream:
			log.Println("Webserver: SSEMsgBroker: closing brokers and connections")
			// close(ws.msgChan)
			// log.Println("Webserver: SSEMsgBroker: Shutdown successful")

			return

		case <-client.Close():
			event := sse.NewSSEBuilder().Event("close").Data("Server closed").Build()

			// fmt.Fprintf(w, "event: close\ndata: server closed\n\n")
			fmt.Fprintf(w, event)
			w.(http.Flusher).Flush()

			ws.connections.Removeclient(client)

			log.Printf("Webserver: EventStream: Client disconnected (ID: %v)", client.ID())
			return

		// case <-r.Context().Done():
		// 	ws.connections.Removeclient(client)
		// 	log.Printf("Webserver: EventStream: Client (ID: %v) disconnected", client.ID())
		// 	return

		case msg := <-ws.msgChan:
			// fmt.Println(msg)
			ws.connections.Broadcast(msg)

		default:
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
