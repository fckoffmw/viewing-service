package room

import (
	"errors"

	"w2g/internal/source"
)

var ErrMaxRoomsReached = errors.New("max rooms reached")
var ErrNotOwner = errors.New("not owner")
var ErrSourceNotFound = errors.New("source not found")

type RoomRepository interface {
	Create(room *Room) error
	GetByInviteCode(inviteCode string) (*Room, error)
	GetByID(id string) (*Room, error)
	Delete(inviteCode string) error
	CountByOwnerID(ownerID string) int
}

type service struct {
	repo            RoomRepository
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

func NewService(r RoomRepository, maxRoomsPerUser int) *service {
	return &service{
		repo:            r,
		maxRoomsPerUser: maxRoomsPerUser,
	}
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
		return nil, err
	}

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
	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		return nil, err
	}

	return &GetResponse{
		ID:            room.ID,
		Name:          room.Name,
		InviteCode:    room.InviteCode,
		OwnerID:       room.OwnerID,
		MembersOnline: 0,
		CreatedAt:     room.CreatedAt,
	}, nil
}

func (s *service) Delete(inviteCode string, userID string) error {
	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		return err
	}

	if room.OwnerID != userID {
		return ErrNotOwner
	}

	return s.repo.Delete(inviteCode)
}

func (s *service) GetRoomByID(id string) (*Room, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetRepo() RoomRepository {
	return s.repo
}

func (s *service) GetCurrentSourceID(roomID string) string {
	return currentSourceID[roomID]
}

func (s *service) SetCurrentSourceID(roomID, sourceID string) {
	currentSourceID[roomID] = sourceID
}