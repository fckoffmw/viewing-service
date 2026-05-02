package chat

import (
	"encoding/json"
	"sync"
)

type RoomHub struct {
	roomID     string
	clients    map[*Client]struct{}
	register  chan *Client
	unregister chan *Client
	broadcast chan []byte
	state     RoomState
	stopCh    chan struct{}
	mu        sync.RWMutex
}

type RoomState struct {
	SourceID  string
	SourceURL string
	Playing   bool
	Position  float64
}

func NewRoomHub(roomID string) *RoomHub {
	return &RoomHub{
		roomID:     roomID,
		clients:    make(map[*Client]struct{}),
		register:  make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		stopCh:    make(chan struct{}),
	}
}

func (h *RoomHub) Run() {
	for {
		select {
		case <-h.stopCh:
			h.closeAllClients()
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (h *RoomHub) Close() {
	close(h.stopCh)
}

func (h *RoomHub) Register() chan<- *Client {
	return h.register
}

func (h *RoomHub) Unregister() chan<- *Client {
	return h.unregister
}

func (h *RoomHub) Broadcast() chan<- []byte {
	return h.broadcast
}

func (h *RoomHub) GetState() (string, string, bool, float64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.state.SourceID, h.state.SourceURL, h.state.Playing, h.state.Position
}

func (h *RoomHub) SetState(sourceID, sourceURL string, playing bool, position float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.state.SourceID = sourceID
	h.state.SourceURL = sourceURL
	h.state.Playing = playing
	h.state.Position = position
}

func (h *RoomHub) BroadcastSourceChanged(sourceID, sourceURL string) {
	msg := map[string]interface{}{
		"type": "source_changed",
		"payload": map[string]string{
			"source_id":  sourceID,
			"source_url": sourceURL,
		},
	}
	data, _ := json.Marshal(msg)

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			delete(h.clients, client)
			close(client.send)
		}
	}
}

func (h *RoomHub) MemberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

func (h *RoomHub) closeAllClients() {
	for client := range h.clients {
		close(client.send)
	}
}