package room

import (
	"errors"
	"fmt"

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

type Service struct {
	repo            RoomRepository
	maxRoomsPerUser int
}

func NewService(r RoomRepository, maxRoomsPerUser int) *Service {
	return &Service{
		repo:            r,
		maxRoomsPerUser: maxRoomsPerUser,
	}
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
	ID             string        `json:"id"`
	Name          string       `json:"name"`
	InviteCode    string       `json:"invite_code"`
	OwnerID     string       `json:"owner_id"`
	MembersOnline int          `json:"members_online"`
	CurrentSource *source.Source `json:"current_source,omitempty"`
	CreatedAt   string       `json:"created_at"`
}

func (s *Service) Create(req CreateRequest, ownerID string) (*CreateResponse, error) {
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

func (s *Service) GetByInviteCode(inviteCode string) (*GetResponse, error) {
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

func (s *Service) Delete(inviteCode string, userID string) error {
	room, err := s.repo.GetByInviteCode(inviteCode)
	if err != nil {
		return err
	}

	if room.OwnerID != userID {
		return ErrNotOwner
	}

	return s.repo.Delete(inviteCode)
}

func (s *Service) GetRoomByID(id string) (*Room, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetRepo() RoomRepository {
	return s.repo
}

var currentSourceID = make(map[string]string)

func (s *Service) GetCurrentSourceID(roomID string) string {
	return currentSourceID[roomID]
}

func (s *Service) SetCurrentSourceID(roomID, sourceID string) {
	currentSourceID[roomID] = sourceID
}

var _ = fmt.Errorf