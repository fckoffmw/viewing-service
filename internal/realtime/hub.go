package realtime

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"
)

type sender interface {
	Send() chan OutgoingMessage
}

type hub struct {
	log        *slog.Logger
	roomID     string
	clients    map[sender]struct{}
	register   chan sender
	unregister chan sender
	incoming   chan incomingEvent
	state      State
	stopCh     chan struct{}
	mu         sync.RWMutex
	onEmpty    func()
}

type State struct {
	SourceID  string
	SourceURL string

	Playing   bool
	Position  float64
	UpdatedAt time.Time
}

type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OutgoingMessage struct {
	Type      string      `json:"type"`
	Username  string      `json:"username,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
}

type ChatPayload struct {
	Text string `json:"text"`
}

type PlayerPayload struct {
	Position float64 `json:"position"`
}

type SyncPayload struct {
	SourceID  string    `json:"source_id"`
	SourceURL string    `json:"source_url"`
	Playing   bool      `json:"playing"`
	Position  float64   `json:"position"`
	UpdatedAt time.Time `json:"updated_at"`
}

type hubManager struct {
	log  *slog.Logger
	hubs map[string]*hub
	mu   sync.RWMutex
}

type HubManager interface {
	GetOrCreate(roomID string) *hub
	GetRoomState(roomID string) State
	GetMembersOnline(roomID string) int
	BroadcastSourceChanged(roomID, sourceID, sourceURL string)
	Remove(roomID string)
}

func newHub(log *slog.Logger, roomID string) *hub {
	return &hub{
		log:        log,
		roomID:     roomID,
		clients:    make(map[sender]struct{}),
		register:   make(chan sender),
		unregister: make(chan sender),
		incoming:   make(chan incomingEvent, 256),
		stopCh:     make(chan struct{}),
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
			h.mu.Lock()
			h.clients[client] = struct{}{}
			syncPayload := OutgoingMessage{
				Type: "sync",
				Payload: SyncPayload{
					SourceID:  h.state.SourceID,
					SourceURL: h.state.SourceURL,
					Playing:   h.state.Playing,
					Position:  h.state.Position,
					UpdatedAt: h.state.UpdatedAt,
				},
			}
			count := len(h.clients)
			h.mu.Unlock()

			client.Send() <- syncPayload
			h.log.Debug("client registered", "room_id", h.roomID, "clients", count)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send())
			}
			shouldShutdown := len(h.clients) == 0
			h.mu.Unlock()

			if shouldShutdown {
				h.shutdownOnEmpty()

				return
			}
		case evt := <-h.incoming:
			h.handleEvent(evt)
		}
	}
}

func (h *hub) handleEvent(evt incomingEvent) {
	switch evt.Message.Type {
	case "chat":
		var payload ChatPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("chat: invalid payload", "err", err)

			return
		}

		payload.Text = strings.TrimSpace(payload.Text)
		if payload.Text == "" {
			return
		}
		if len(payload.Text) > maxTextLen {
			payload.Text = payload.Text[:maxTextLen]
		}

		outgoing := OutgoingMessage{
			Type:     "chat",
			Username: evt.Username,
			Payload:  payload,
		}

		h.mu.Lock()
		for client := range h.clients {
			if client == evt.Sender {
				continue
			}

			select {
			case client.Send() <- outgoing:
			default:
				delete(h.clients, client)
				close(client.Send())
			}
		}
		shouldShutdown := len(h.clients) == 0
		h.mu.Unlock()

		if shouldShutdown {
			h.shutdownOnEmpty()
		}

	case "play":
		var payload PlayerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("play: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.state.Playing = true
		h.state.Position = payload.Position
		h.state.UpdatedAt = time.Now()
		outgoing := OutgoingMessage{
			Type:     "play",
			Username: evt.Username,
			Payload:  payload,
		}
		h.mu.Unlock()

		h.broadcastAll(outgoing)

	case "pause":
		var payload PlayerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("pause: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.state.Playing = false
		h.state.Position = payload.Position
		h.state.UpdatedAt = time.Now()
		outgoing := OutgoingMessage{
			Type:     "pause",
			Username: evt.Username,
			Payload:  payload,
		}
		h.mu.Unlock()

		h.broadcastAll(outgoing)

	case "seek":
		var payload PlayerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("seek: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.state.Position = payload.Position
		h.state.UpdatedAt = time.Now()
		outgoing := OutgoingMessage{
			Type:     "seek",
			Username: evt.Username,
			Payload:  payload,
		}
		h.mu.Unlock()

		h.broadcastAll(outgoing)

	case "source_changed":
		var payload struct {
			SourceID  string `json:"source_id"`
			SourceURL string `json:"source_url"`
		}
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("source_changed: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.state.SourceID = payload.SourceID
		h.state.SourceURL = payload.SourceURL
		h.state.UpdatedAt = time.Now()
		outgoing := OutgoingMessage{
			Type:    "source_changed",
			Payload: payload,
		}
		h.mu.Unlock()

		h.broadcastAll(outgoing)
	}
}

func (h *hub) broadcastAll(msg OutgoingMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range h.clients {
		select {
		case c.Send() <- msg:
		default:
			delete(h.clients, c)
			close(c.Send())
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

func (h *hub) Incoming() chan<- incomingEvent {
	return h.incoming
}

func (h *hub) GetState() State {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.state
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
	payload, _ := json.Marshal(map[string]string{
		"source_id":  sourceID,
		"source_url": sourceURL,
	})

	h.incoming <- incomingEvent{
		Message: IncomingMessage{
			Type:    "source_changed",
			Payload: payload,
		},
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
	h.mu.Lock()
	defer h.mu.Unlock()

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

func (m *hubManager) GetRoomState(roomID string) State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub.GetState()
	}
	return State{}
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
