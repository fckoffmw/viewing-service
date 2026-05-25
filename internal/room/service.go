package room

import (
	"errors"
	"fmt"
	"log/slog"
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
	log             *slog.Logger
	repo            Repository
	sourceGetter    SourceGetter
	hub             HubManager
	maxRoomsPerUser int

	currentSourceID map[string]string
	mu              sync.RWMutex
}

func NewService(log *slog.Logger, r Repository, sg SourceGetter, h HubManager, maxRoomsPerUser int) *service {
	return &service{
		log:             log,
		repo:            r,
		sourceGetter:    sg,
		hub:             h,
		maxRoomsPerUser: maxRoomsPerUser,
		currentSourceID: make(map[string]string),
	}
}

func (s *service) GetAll() []GetResponse {
	s.log.Debug("getting all rooms")
	rooms := s.repo.GetAll()
	s.log.Debug("rooms fetched", "count", len(rooms))

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
	s.log.Debug("creating room", "name", req.Name, "owner_id", ownerID)

	if s.maxRoomsPerUser > 0 && s.repo.CountByOwnerID(ownerID) >= s.maxRoomsPerUser {
		s.log.Warn("max rooms reached", "owner_id", ownerID, "max", s.maxRoomsPerUser)

		return nil, ErrMaxRoomsReached
	}

	room := &Room{
		Name:    req.Name,
		OwnerID: ownerID,
	}

	if err := s.repo.Create(room); err != nil {
		s.log.Error("failed to create room", "err", err, "owner_id", ownerID)

		return nil, err
	}

	s.log.Info("room created", "room_id", room.ID, "invite_code", room.InviteCode, "owner_id", ownerID)

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
	s.log.Debug("getting room by invite code", "invite_code", inviteCode)

	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		s.log.Error("room not found", "invite_code", inviteCode, "err", err)

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
			s.log.Debug("current source loaded", "source_id", sourceID, "source_name", src.Name)
		}
	}

	return resp, nil
}

func (s *service) Delete(inviteCode string, userID string) error {
	s.log.Debug("deleting room", "invite_code", inviteCode, "user_id", userID)

	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		s.log.Error("failed to get room for delete", "invite_code", inviteCode, "err", err)

		return err
	}

	if room.OwnerID != userID {
		s.log.Warn("delete denied - not owner", "invite_code", inviteCode, "user_id", userID, "owner_id", room.OwnerID)

		return ErrNotOwner
	}

	if err := s.repo.Delete(inviteCode); err != nil {
		s.log.Error("failed to delete room", "invite_code", inviteCode, "err", err)

		return err
	}

	s.hub.Remove(inviteCode)

	s.log.Info("room deleted", "invite_code", inviteCode, "room_id", room.ID)

	return nil
}

func (s *service) PatchSource(inviteCode, ownerID, sourceID string) error {
	s.log.Debug("patching room source", "invite_code", inviteCode, "source_id", sourceID)

	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		s.log.Error("failed to get room for patch", "invite_code", inviteCode, "err", err)

		return fmt.Errorf("get room %s: %w", inviteCode, err)
	}

	if room.OwnerID != ownerID {
		s.log.Warn("patch source denied - not owner", "invite_code", inviteCode, "user_id", ownerID, "owner_id", room.OwnerID)

		return ErrNotOwner
	}

	src, err := s.sourceGetter.GetSourceByID(sourceID)
	if err != nil {
		s.log.Warn("source not found", "source_id", sourceID, "err", err)

		return ErrSourceNotFound
	}

	s.setCurrentSourceID(room.ID, sourceID)
	s.hub.BroadcastSourceChanged(inviteCode, sourceID, src.URL)

	s.log.Info("source changed", "room_id", room.ID, "invite_code", inviteCode, "source_id", sourceID)

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
