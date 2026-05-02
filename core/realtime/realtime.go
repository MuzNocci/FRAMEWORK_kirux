package realtime

import (
	"kyrux/core/events"
	"net/http"
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
	bus     *events.Bus
}

func NewHub(bus *events.Bus) *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		bus:     bus,
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c.id] = c
	h.mu.Unlock()
}

func (h *Hub) Unregister(id string) {
	h.mu.Lock()
	if c, ok := h.clients[id]; ok {
		c.close()
		delete(h.clients, id)
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(event string, payload any) {
	h.bus.Publish(event, payload)
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := newClient(w, r, h)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}
	h.Register(c)
	go c.writePump()
	c.readPump()
}
