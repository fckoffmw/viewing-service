package chat

import "sync"

type HubManager struct {
	hubs map[string]*RoomHub
	mu   sync.RWMutex
}

func NewHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*RoomHub),
	}
}

func (m *HubManager) GetOrCreate(roomID string) *RoomHub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub
	}

	hub := NewRoomHub(roomID)
	m.hubs[roomID] = hub
	go hub.Run()

	return hub
}

func (m *HubManager) Get(roomID string) (*RoomHub, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hub, ok := m.hubs[roomID]
	return hub, ok
}

func (m *HubManager) Remove(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[roomID]; ok {
		hub.Close()
		delete(m.hubs, roomID)
	}
}

func (m *HubManager) GetRoomState(roomID string) (string, string, bool, float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub.GetState()
	}
	return "", "", false, 0
}

func (m *HubManager) GetMembersOnline(roomID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub.MemberCount()
	}
	return 0
}

func (m *HubManager) BroadcastSourceChanged(roomID, sourceID, sourceURL string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hub, ok := m.hubs[roomID]; ok {
		hub.BroadcastSourceChanged(sourceID, sourceURL)
	}
}