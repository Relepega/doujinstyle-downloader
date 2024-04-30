package ssehub

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	id     int64
	writer http.ResponseWriter
}

type clients []*Client

type ConnectionsHub struct {
	sync.Mutex

	clients clients
}

func (h *ConnectionsHub) AddClient(w http.ResponseWriter) *Client {
	h.Lock()
	defer h.Unlock()

	newClient := &Client{
		id:     time.Now().UnixMicro(),
		writer: w,
	}

	h.clients = append(h.clients, newClient)

	return newClient
}

func (h *ConnectionsHub) Removeclient(c *Client) {
	h.Lock()
	defer h.Unlock()

	var filtered clients

	for _, client := range h.clients {
		if client.id == c.id {
			continue
		}

		filtered = append(filtered, client)
	}

	h.clients = filtered
}

func (h *ConnectionsHub) Broadcast(msg string) {
	h.Lock()
	defer h.Unlock()

	for _, c := range h.clients {
		fmt.Fprintf(c.writer, msg)
		c.writer.(http.Flusher).Flush()
	}
}
