package chat

import (
	"encoding/json"
	"log/slog"
	"sync"
)

type sender interface {
	Send() chan []byte
}

type hub struct {
	log       *slog.Logger
	roomID    string
	clients   map[sender]struct{}
	register  chan sender
	unregister chan sender
	broadcast  chan []byte
	state      state
	stopCh     chan struct{}
	mu         sync.RWMutex
	onEmpty   func()
}

type state struct {
	SourceID  string
	SourceURL string
	Playing   bool
	Position  float64
}

type hubManager struct {
	log  *slog.Logger
	hubs map[string]*hub
	mu   sync.RWMutex
}

type HubManager interface {
	GetOrCreate(roomID string) *hub
	GetRoomState(roomID string) (string, string, bool, float64)
	GetMembersOnline(roomID string) int
	BroadcastSourceChanged(roomID, sourceID, sourceURL string)
	Remove(roomID string)
}

func newHub(log *slog.Logger, roomID string) *hub {
	return &hub{
		log:       log,
		roomID:    roomID,
		clients:   make(map[sender]struct{}),
		register:  make(chan sender),
		unregister: make(chan sender),
		broadcast:  make(chan []byte, 256),
		stopCh:    make(chan struct{}),
	}
}

func NewHubManager(log *slog.Logger) HubManager {
	return &hubManager{
		log:  log,
		hubs: make(map[string]*hub),
	}
}

func (h *hub) Run() {
	h.log.Debug("hub started", "room_id", h.roomID)
	for {
		select {
		case <-h.stopCh:
			h.log.Debug("hub stopped", "room_id", h.roomID)
			h.closeAllClients()
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
			h.log.Debug("client registered", "room_id", h.roomID, "clients", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send())
				h.log.Debug("client unregistered", "room_id", h.roomID, "clients", len(h.clients))
			}
			if len(h.clients) == 0 {
				h.shutdownOnEmpty()

				return
			}
		case msg := <-h.broadcast:
			h.log.Debug("broadcasting message", "room_id", h.roomID, "clients", len(h.clients), "msg_len", len(msg))
			for client := range h.clients {
				select {
				case client.Send() <- msg:
				default:
					delete(h.clients, client)
					close(client.Send())
				}
			}
			if len(h.clients) == 0 {
				h.shutdownOnEmpty()

				return
			}
		}
	}
}

func (h *hub) Close() {
	close(h.stopCh)
}

func (h *hub) Register() chan<- sender {
	return h.register
}

func (h *hub) Unregister() chan<- sender {
	return h.unregister
}

func (h *hub) Broadcast() chan<- []byte {
	return h.broadcast
}

func (h *hub) GetState() (string, string, bool, float64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.state.SourceID, h.state.SourceURL, h.state.Playing, h.state.Position
}

func (h *hub) SetState(sourceID, sourceURL string, playing bool, position float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.state.SourceID = sourceID
	h.state.SourceURL = sourceURL
	h.state.Playing = playing
	h.state.Position = position
}

func (h *hub) BroadcastSourceChanged(sourceID, sourceURL string) {
	_, _, playing, position := h.GetState()
	h.SetState(sourceID, sourceURL, playing, position)

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
		case client.Send() <- data:
		default:
			delete(h.clients, client)
			close(client.Send())
		}
	}
}

func (h *hub) MemberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

func (h *hub) shutdownOnEmpty() {
	h.log.Debug("hub empty, stopping", "room_id", h.roomID)
	if h.onEmpty != nil {
		h.onEmpty()
	}
}

func (h *hub) closeAllClients() {
	for client := range h.clients {
		close(client.Send())
	}
}

func (m *hubManager) GetOrCreate(roomID string) *hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub
	}

	hub := newHub(m.log, roomID)
	hub.onEmpty = func() {
		m.Remove(roomID)
	}
	m.hubs[roomID] = hub
	go hub.Run()

	return hub
}

func (m *hubManager) Get(roomID string) (*hub, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hub, ok := m.hubs[roomID]
	return hub, ok
}

func (m *hubManager) Remove(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[roomID]; ok {
		hub.Close()
		delete(m.hubs, roomID)
	}
}

func (m *hubManager) GetRoomState(roomID string) (string, string, bool, float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub.GetState()
	}
	return "", "", false, 0
}

func (m *hubManager) GetMembersOnline(roomID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub.MemberCount()
	}
	return 0
}

func (m *hubManager) BroadcastSourceChanged(roomID, sourceID, sourceURL string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		hub.BroadcastSourceChanged(sourceID, sourceURL)
	}
}