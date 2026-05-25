package room

import (
	"fmt"
	"log/slog"
	"sync"
	"w2g/internal/strutils"
)

type store struct {
	log     *slog.Logger
	rooms   map[string]*Room
	storage Storage
	mu      sync.RWMutex
}

type Storage interface {
	GetAllRooms() ([]Room, error)
	AddRoom(r *Room) (string, error)
	UpdateRoom(r Room) error
	DeleteRoom(id string) error
}

func NewStore(log *slog.Logger, storage Storage) (*store, error) {
	s := &store{
		log:     log,
		rooms:   make(map[string]*Room),
		storage: storage,
	}

	if err := s.loadFromCSV(); err != nil {
		log.Error("failed to load rooms from storage", "err", err)
		return nil, fmt.Errorf("load rooms from storage: %w", err)
	}

	log.Info("room store initialized", "rooms_count", len(s.rooms))
	return s, nil
}

func (s *store) loadFromCSV() error {
	rooms, err := s.storage.GetAllRooms()
	if err != nil {
		return fmt.Errorf("when load rooms: %w", err)
	}

	for i := range rooms {
		s.rooms[rooms[i].InviteCode] = &rooms[i]
	}

	return nil
}

func (s *store) GetAll() []*Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		result = append(result, room)
	}
	return result
}

func (s *store) Create(room *Room) error {
	inviteCode, err := strutils.GenerateInviteCode()
	if err != nil {
		return err
	}
	room.InviteCode = inviteCode

	id, err := s.storage.AddRoom(room)
	if err != nil {
		return err
	}

	room.ID = id
	s.rooms[room.InviteCode] = room

	return nil
}

func (s *store) GetByInviteCode(inviteCode string) (*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[inviteCode]
	if !ok {
		return nil, fmt.Errorf("room not found")
	}
	return room, nil
}

func (s *store) GetByID(id string) (*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, room := range s.rooms {
		if room.ID == id {
			return room, nil
		}
	}
	return nil, fmt.Errorf("room not found")
}

func (s *store) GetByOwnerID(ownerID string) ([]*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rooms []*Room
	for _, room := range s.rooms {
		if room.OwnerID == ownerID {
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}

func (s *store) Delete(inviteCode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[inviteCode]
	if !ok {
		return fmt.Errorf("room not found")
	}

	delete(s.rooms, inviteCode)

	return s.storage.DeleteRoom(room.ID)
}

func (s *store) CountByOwnerID(ownerID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, room := range s.rooms {
		if room.OwnerID == ownerID {
			count++
		}
	}
	return count
}
