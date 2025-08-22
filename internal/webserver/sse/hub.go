package sse

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	id      int64
	writer  http.ResponseWriter
	request *http.Request
	close   chan struct{}
}

func (c *Client) Close() <-chan struct{} { return c.close }

func (c *Client) ID() int64 { return c.id }

type clients []*Client

type Hub struct {
	sync.Mutex

	clients clients
	open    bool
}

func NewHub() *Hub {
	return &Hub{
		open: true,
	}
}

func (h *Hub) AddClient(w http.ResponseWriter, r *http.Request) *Client {
	h.Lock()
	defer h.Unlock()

	newClient := &Client{
		id:      time.Now().UnixMicro(),
		writer:  w,
		request: r,
		close:   make(chan struct{}),
	}

	h.clients = append(h.clients, newClient)

	return newClient
}

func (h *Hub) Removeclient(c *Client) {
	h.Lock()
	defer h.Unlock()

	if !h.open {
		return
	}

	var filtered clients

	for _, client := range h.clients {
		if client.id == c.id {
			close(client.close)
			client.request.Body.Close()

			continue
		}

		filtered = append(filtered, client)
	}

	h.clients = filtered
}

func (h *Hub) Broadcast(msg string) {
	h.Lock()
	defer h.Unlock()

	if !h.open {
		return
	}

	for _, c := range h.clients {
		fmt.Fprintf(c.writer, "%v", msg)
		c.writer.(http.Flusher).Flush()
	}
}

func (h *Hub) Shutdown() {
	h.Lock()
	defer h.Unlock()

	if !h.open {
		return
	}

	log.Println("Webserver: ConnectionsHub: shutting down connected clients")

	// for _, c := range h.clients {
	// 	c.close <- struct{}{}
	// 	close(c.close)
	//
	// 	c.request.Body.Close()
	// }

	h.open = false

	log.Println("Webserver: ConnectionsHub: shutdown complete")
}
