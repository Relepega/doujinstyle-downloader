package sse

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

type Hub struct {
	sync.Mutex

	clients clients
}

func (h *Hub) AddClient(w http.ResponseWriter) *Client {
	h.Lock()
	defer h.Unlock()

	newClient := &Client{
		id:     time.Now().UnixMicro(),
		writer: w,
	}

	h.clients = append(h.clients, newClient)

	return newClient
}

func (h *Hub) Removeclient(c *Client) {
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

func (h *Hub) Broadcast(msg string) {
	h.Lock()
	defer h.Unlock()

	for _, c := range h.clients {
		fmt.Fprintf(c.writer, "%v", msg)
		c.writer.(http.Flusher).Flush()
	}
}
