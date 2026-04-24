package sync

import (
	"log/slog"
)

type Hub struct {
	clients    map[*wsClient]struct{}
	register   chan registerRequest
	unregister chan *wsClient
	Ch         chan []byte
	Log        *slog.Logger
}

type wsClient struct {
	Send chan []byte
}

type registerRequest struct {
	client   *wsClient
	accepted chan bool
}

func (h *Hub) Broadcast() chan []byte {
	return h.Ch
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*wsClient]struct{}),
		register:   make(chan registerRequest),
		unregister: make(chan *wsClient),
		Ch:         make(chan []byte),
		Log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case req := <-h.register:
			client := req.client
			h.clients[client] = struct{}{}
			req.accepted <- true
			h.Log.Debug("sync client registered", "clients", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.Log.Debug("sync client unregistered", "clients", len(h.clients))
		case msg := <-h.Ch:
			h.Log.Debug("sync broadcast", "clients", len(h.clients), "msgLen", len(msg))
			for client := range h.clients {
				select {
				case client.Send <- msg:
				default:
					delete(h.clients, client)
					close(client.Send)
				}
			}
		}
	}
}
