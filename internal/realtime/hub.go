package realtime

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const (
	MsgTypeSync          = "sync"
	MsgTypeChat          = "chat"
	MsgTypePlay          = "play"
	MsgTypePause         = "pause"
	MsgTypeSeek          = "seek"
	MsgTypeSourceChanged = "source_changed"
	MsgTypeSticker       = "sticker"

	keySourceID  = "source_id"
	keySourceURL = "source_url"

	incomingBufSize = 256
)

type sender interface {
	Send() chan outgoingMessage
}

type hub struct {
	log        *slog.Logger
	roomID     string
	clients    map[sender]struct{}
	register   chan sender
	unregister chan sender
	incoming   chan incomingEvent
	state      state
	stopCh     chan struct{}
	mu         sync.RWMutex
	onEmpty    func()
}

type state struct {
	SourceID  string
	SourceURL string

	Playing   bool
	Position  float64
	UpdatedAt time.Time
}

type incomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type outgoingMessage struct {
	Type      string      `json:"type"`
	Username  string      `json:"username,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
}

type hubManager struct {
	log  *slog.Logger
	hubs map[string]*hub
	mu   sync.RWMutex
}

func newHub(log *slog.Logger, roomID string) *hub {
	return &hub{
		log:        log,
		roomID:     roomID,
		clients:    make(map[sender]struct{}),
		register:   make(chan sender),
		unregister: make(chan sender),
		incoming:   make(chan incomingEvent, incomingBufSize),
		stopCh:     make(chan struct{}),
	}
}

func NewHubManager(log *slog.Logger) *hubManager {
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
			syncPayload := outgoingMessage{
				Type:    MsgTypeSync,
				Payload: newSyncPayload(h.state),
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
	h.log.Debug("event received",
		"type", evt.Message.Type,
		"username", evt.Username,
		"payload", string(evt.Message.Payload),
	)

	switch evt.Message.Type {
	case MsgTypeChat:
		var payload chatPayload
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

		outgoing := outgoingMessage{
			Type:     MsgTypeChat,
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
				h.log.Warn("chat: dropping slow client", "username", evt.Username)
				delete(h.clients, client)
				close(client.Send())
			}
		}
		shouldShutdown := len(h.clients) == 0
		h.mu.Unlock()

		h.log.Debug("chat broadcast done", "clients_remaining", len(h.clients))

		if shouldShutdown {
			h.shutdownOnEmpty()
		}

	case MsgTypePlay:
		var payload playerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("play: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.applyPlayerEvent(true, payload.Position)
		h.mu.Unlock()

		h.log.Debug("player state after play", "playing", h.getState().Playing, "position", h.getState().Position)

		h.broadcastAll(outgoingMessage{
			Type:     MsgTypePlay,
			Username: evt.Username,
			Payload:  payload,
		})

	case MsgTypePause:
		var payload playerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("pause: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.applyPlayerEvent(false, payload.Position)
		h.mu.Unlock()

		h.log.Debug("player state after pause", "playing", h.getState().Playing, "position", h.getState().Position)

		h.broadcastAll(outgoingMessage{
			Type:     MsgTypePause,
			Username: evt.Username,
			Payload:  payload,
		})

	case MsgTypeSeek:
		var payload playerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("seek: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.state.Position = payload.Position
		if payload.Playing != nil {
			h.state.Playing = *payload.Playing
		}
		h.state.UpdatedAt = time.Now()
		h.mu.Unlock()

		h.log.Debug("player state after seek", "playing", h.getState().Playing, "position", h.getState().Position)

		h.broadcastAll(outgoingMessage{
			Type:     MsgTypeSeek,
			Username: evt.Username,
			Payload:  payload,
		})

	case MsgTypeSourceChanged:
		var payload sourceChangedPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("source_changed: invalid payload", "err", err)

			return
		}

		h.mu.Lock()
		h.applySourceEvent(payload.SourceID, payload.SourceURL)
		h.mu.Unlock()

		h.log.Debug("player state after source_changed", "source_id", payload.SourceID, "source_url", payload.SourceURL)

		h.broadcastAll(outgoingMessage{
			Type:    MsgTypeSourceChanged,
			Payload: payload,
		})

	case MsgTypeSticker:
		var payload stickerPayload
		if err := json.Unmarshal(evt.Message.Payload, &payload); err != nil {
			h.log.Debug("sticker: invalid payload", "err", err)

			return
		}

		payload.StickerID = strings.TrimSpace(payload.StickerID)
		if payload.StickerID == "" {

			return
		}

		outgoing := outgoingMessage{
			Type:     MsgTypeSticker,
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
				h.log.Warn("sticker: dropping slow client", "username", evt.Username)
				delete(h.clients, client)
				close(client.Send())
			}
		}
		shouldShutdown := len(h.clients) == 0
		h.mu.Unlock()

		h.log.Debug("sticker broadcast done", "clients_remaining", len(h.clients))

		if shouldShutdown {
			h.shutdownOnEmpty()
		}
	}
}

func (h *hub) applyPlayerEvent(playing bool, position float64) {
	h.state.Playing = playing
	h.state.Position = position
	h.state.UpdatedAt = time.Now()
}

func (h *hub) applySourceEvent(sourceID, sourceURL string) {
	h.state.SourceID = sourceID
	h.state.SourceURL = sourceURL
	h.state.UpdatedAt = time.Now()
}

func (h *hub) broadcastAll(msg outgoingMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.log.Debug("broadcasting",
		"type", msg.Type,
		"username", msg.Username,
		"clients", len(h.clients),
	)

	for c := range h.clients {
		select {
		case c.Send() <- msg:
		default:
			h.log.Warn("broadcast: dropping slow client", "type", msg.Type)
			delete(h.clients, c)
			close(c.Send())
		}
	}

	h.log.Debug("broadcast done", "type", msg.Type, "clients_after", len(h.clients))
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

func (h *hub) getState() state {
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
		keySourceID:  sourceID,
		keySourceURL: sourceURL,
	})

	evt := incomingEvent{
		Message: incomingMessage{
			Type:    MsgTypeSourceChanged,
			Payload: payload,
		},
	}

	select {
	case h.incoming <- evt:
	default:
		h.log.Warn("broadcast source changed: incoming channel full, dropping event", "room_id", h.roomID)
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

func (m *hubManager) GetRoomState(roomID string) state {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub.getState()
	}
	return state{}
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
