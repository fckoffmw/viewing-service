package room

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"w2g/internal/source"
)

var ErrMaxRoomsReached = errors.New("max rooms reached")
var ErrNotOwner = errors.New("not owner")
var ErrSourceNotFound = errors.New("source not found")

type Repository interface {
	Create(room *Room) error
	GetByInviteCode(inviteCode string) (*Room, error)
	GetByID(id string) (*Room, error)
	Delete(inviteCode string) error
	CountByOwnerID(ownerID string) int
	GetAll() []*Room
}

type SourceGetter interface {
	GetSourceByID(id string) (*source.Source, error)
}

type HubManager interface {
	GetMembersOnline(roomID string) int
	BroadcastSourceChanged(roomID, sourceID, sourceURL string)
	Remove(roomID string)
}

type service struct {
	repo            Repository
	sourceGetter    SourceGetter
	hub             HubManager
	maxRoomsPerUser int

	currentSourceID map[string]string
	mu              sync.RWMutex
}

func NewService(r Repository, sg SourceGetter, h HubManager, maxRoomsPerUser int) *service {
	return &service{
		repo:            r,
		sourceGetter:    sg,
		hub:             h,
		maxRoomsPerUser: maxRoomsPerUser,
		currentSourceID: make(map[string]string),
	}
}

func (s *service) GetAll() []GetResponse {
	rooms := s.repo.GetAll()

	result := make([]GetResponse, 0, len(rooms))

	for _, room := range rooms {
		resp := GetResponse{
			ID:         room.ID,
			Name:       room.Name,
			InviteCode: room.InviteCode,
			OwnerID:    room.OwnerID,
			CreatedAt:  room.CreatedAt.Format(time.RFC3339),
		}

		sourceID := s.getCurrentSourceID(room.ID)
		if sourceID != "" {
			src, err := s.sourceGetter.GetSourceByID(sourceID)
			if err == nil && src != nil {
				resp.CurrentSource = src
			}
		}

		resp.MembersOnline = s.hub.GetMembersOnline(room.InviteCode)

		result = append(result, resp)
	}

	return result
}

func (s *service) Create(req CreateRequest, ownerID string) (*CreateResponse, error) {
	if s.maxRoomsPerUser > 0 && s.repo.CountByOwnerID(ownerID) >= s.maxRoomsPerUser {
		return nil, ErrMaxRoomsReached
	}

	room := &Room{
		Name:    req.Name,
		OwnerID: ownerID,
	}

	if err := s.repo.Create(room); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	return &CreateResponse{
		ID:         room.ID,
		Name:       room.Name,
		InviteCode: room.InviteCode,
		InviteURL:  "/room/" + room.InviteCode,
		OwnerID:    room.OwnerID,
		CreatedAt:  room.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) GetByInviteCode(inviteCode string) (*GetResponse, error) {
	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		return nil, fmt.Errorf("get room %s: %w", inviteCode, err)
	}

	membersOnline := s.hub.GetMembersOnline(inviteCode)

	resp := &GetResponse{
		ID:            room.ID,
		Name:          room.Name,
		InviteCode:    room.InviteCode,
		OwnerID:       room.OwnerID,
		MembersOnline: membersOnline,
		CreatedAt:     room.CreatedAt.Format(time.RFC3339),
	}

	sourceID := s.getCurrentSourceID(room.ID)
	if sourceID != "" {
		src, err := s.sourceGetter.GetSourceByID(sourceID)
		if err == nil && src != nil {
			resp.CurrentSource = src
		}
	}

	return resp, nil
}

func (s *service) Delete(inviteCode string, userID string) error {
	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		return fmt.Errorf("get room %s: %w", inviteCode, err)
	}

	if room.OwnerID != userID {
		return ErrNotOwner
	}

	if err := s.repo.Delete(inviteCode); err != nil {
		return fmt.Errorf("delete room %s: %w", inviteCode, err)
	}

	s.hub.Remove(inviteCode)

	return nil
}

func (s *service) PatchSource(inviteCode, ownerID, sourceID string) error {
	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		return fmt.Errorf("get room %s: %w", inviteCode, err)
	}

	if room.OwnerID != ownerID {
		return ErrNotOwner
	}

	src, err := s.sourceGetter.GetSourceByID(sourceID)
	if err != nil {
		return ErrSourceNotFound
	}

	s.setCurrentSourceID(room.ID, sourceID)
	s.hub.BroadcastSourceChanged(inviteCode, sourceID, src.URL)

	return nil
}

func (s *service) GetByID(id string) (*Room, error) {
	return s.repo.GetByID(id)
}

func (s *service) getCurrentSourceID(roomID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentSourceID[roomID]
}

func (s *service) setCurrentSourceID(roomID, sourceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentSourceID[roomID] = sourceID
}
