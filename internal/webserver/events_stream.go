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

	client := ws.connections.AddClient(w, r)
	log.Println("Webserver: EventStream: New client connected")

	for {
		select {
		case <-ws.ctx.Done():
			log.Println("Webserver: EventStream: Context cancellation received")
			r.Body.Close()
			ws.connections.Removeclient(client)
			log.Println("Webserver: EventStream: Closing stream due to server shutdown")
			return

		case <-r.Context().Done():
			ws.connections.Removeclient(client)
			log.Println("Webserver: EventStream: Client disconnected")
			return

		case msg, open := <-ws.msgChan:
			if !open {
				log.Println("Webserver: EventStream: Channel closed due to server shutdown")
				r.Body.Close()
			}

			// fmt.Println(s)
			ws.connections.Broadcast(msg)

		default:
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
