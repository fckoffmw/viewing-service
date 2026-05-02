package room

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
)

type Store struct {
	rooms      map[string]*Room
	csvStorage CSVStorage
	mu         sync.RWMutex
}

type CSVStorage interface {
	GetAllRooms() ([]Room, error)
	AddRoom(r Room) (string, error)
	UpdateRoom(r Room) error
	DeleteRoom(id string) error
}

func NewStore(csvStorage CSVStorage) (*Store, error) {
	s := &Store{
		rooms:      make(map[string]*Room),
		csvStorage: csvStorage,
	}

	if err := s.loadFromCSV(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) loadFromCSV() error {
	rooms, err := s.csvStorage.GetAllRooms()
	if err != nil {
		return fmt.Errorf("when load rooms: %w", err)
	}

	for i := range rooms {
		s.rooms[rooms[i].InviteCode] = &rooms[i]
	}

	return nil
}

func (s *Store) Create(room *Room) error {
	// Generate invite code before saving
	inviteCode, err := generateInviteCode()
	if err != nil {
		return err
	}
	room.InviteCode = inviteCode

	id, err := s.csvStorage.AddRoom(*room)
	if err != nil {
		return err
	}

	room.ID = id
	s.rooms[room.InviteCode] = room

	return nil
}

func (s *Store) GetByInviteCode(inviteCode string) (*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[inviteCode]
	if !ok {
		return nil, fmt.Errorf("room not found")
	}
	return room, nil
}

func (s *Store) GetByID(id string) (*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, room := range s.rooms {
		if room.ID == id {
			return room, nil
		}
	}
	return nil, fmt.Errorf("room not found")
}

func (s *Store) GetByOwnerID(ownerID string) ([]*Room, error) {
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

func (s *Store) Delete(inviteCode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[inviteCode]
	if !ok {
		return fmt.Errorf("room not found")
	}

	delete(s.rooms, inviteCode)

	return s.csvStorage.DeleteRoom(room.ID)
}

func (s *Store) CountByOwnerID(ownerID string) int {
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

func generateInviteCode() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}

	return string(result), nil
}
