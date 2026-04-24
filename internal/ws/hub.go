package ws

import (
	"log/slog"
)

type Hub struct {
	clients    map[*Client]struct{}
	register   chan registerRequest
	unregister chan *Client
	broadcast  chan []byte
	maxClients int
	log        *slog.Logger
}

type registerRequest struct {
	client   *Client
	accepted chan bool
}

func NewHub(maxClients int, log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan registerRequest),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		maxClients: maxClients,
		log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case req := <-h.register:
			client := req.client
			if len(h.clients) >= h.maxClients {
				h.log.Debug("ws register rejected", "reason", "room full", "current", len(h.clients), "max", h.maxClients)
				req.accepted <- false
				continue
			}
			h.clients[client] = struct{}{}
			req.accepted <- true
			h.log.Debug("ws client registered", "clients", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.log.Debug("ws client unregistered", "clients", len(h.clients))
		case msg := <-h.broadcast:
			h.log.Debug("ws broadcast", "clients", len(h.clients), "msgLen", len(msg))
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
