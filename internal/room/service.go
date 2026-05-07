package room

import (
	"errors"
	"fmt"
	"log/slog"

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
	GetSourceById(id string) (*source.Source, error)
}

type HubGetter interface {
	GetMembersOnline(roomID string) int
}

type service struct {
	log             *slog.Logger
	repo            Repository
	sourceGetter    SourceGetter
	hubGetter       HubGetter
	maxRoomsPerUser int
}

type CreateRequest struct {
	Name string `json:"name"`
}

type CreateResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
	InviteURL  string `json:"invite_url"`
	OwnerID    string `json:"owner_id"`
	CreatedAt  string `json:"created_at"`
}

type GetResponse struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	InviteCode     string          `json:"invite_code"`
	OwnerID        string          `json:"owner_id"`
	MembersOnline  int             `json:"members_online"`
	CurrentSource  *source.Source  `json:"current_source,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

var currentSourceID = make(map[string]string)

func NewService(log *slog.Logger, r Repository, sg SourceGetter, hg HubGetter, maxRoomsPerUser int) *service {
	return &service{
		log:             log,
		repo:            r,
		sourceGetter:    sg,
		hubGetter:       hg,
		maxRoomsPerUser: maxRoomsPerUser,
	}
}

func (s *service) GetAllRooms() []GetResponse {
	s.log.Debug("getting all rooms")
	rooms := s.repo.GetAll()
	s.log.Debug("rooms fetched", "count", len(rooms))

	result := make([]GetResponse, 0, len(rooms))

	for _, room := range rooms {
		resp := GetResponse{
			ID:            room.ID,
			Name:          room.Name,
			InviteCode:    room.InviteCode,
			OwnerID:       room.OwnerID,
			MembersOnline: 0,
			CreatedAt:     room.CreatedAt,
		}

		sourceID := currentSourceID[room.ID]
		if sourceID != "" {
			src, err := s.sourceGetter.GetSourceById(sourceID)
			if err == nil && src != nil {
				resp.CurrentSource = src
			}
		}

		if s.hubGetter != nil {
			resp.MembersOnline = s.hubGetter.GetMembersOnline(room.InviteCode)
		}

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
		CreatedAt:  room.CreatedAt,
	}, nil
}

func (s *service) GetByInviteCode(inviteCode string) (*GetResponse, error) {
	s.log.Debug("getting room by invite code", "invite_code", inviteCode)

	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		s.log.Error("room not found", "invite_code", inviteCode, "err", err)
		return nil, fmt.Errorf("get room %s: %w", inviteCode, err)
	}

	resp := &GetResponse{
		ID:            room.ID,
		Name:          room.Name,
		InviteCode:    room.InviteCode,
		OwnerID:       room.OwnerID,
		MembersOnline: 0,
		CreatedAt:     room.CreatedAt,
	}

	sourceID := currentSourceID[room.ID]
	if sourceID != "" {
		src, err := s.sourceGetter.GetSourceById(sourceID)
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

	s.log.Info("room deleted", "invite_code", inviteCode, "room_id", room.ID)
	return nil
}

func (s *service) GetRoomByID(id string) (*Room, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetRepo() Repository {
	return s.repo
}

func (s *service) GetCurrentSourceID(roomID string) string {
	return currentSourceID[roomID]
}

func (s *service) SetCurrentSourceID(roomID, sourceID string) {
	s.log.Debug("setting current source", "room_id", roomID, "source_id", sourceID)
	currentSourceID[roomID] = sourceID
}