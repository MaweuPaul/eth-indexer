package hub

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan *Event
	mu        sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan *Event),
	}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
}

func (h *Hub) Broadcast(event *Event) {
	h.broadcast <- event
}

func (h *Hub) Run() {

	for event := range h.broadcast {
		h.mu.Lock()

		for conn := range h.clients {
			if err := conn.WriteJSON(event); err != nil {
				log.Println("Failed to send event to client:", err)
				conn.Close()
				delete(h.clients, conn)
			}
		}
		h.mu.Unlock()
	}
}
