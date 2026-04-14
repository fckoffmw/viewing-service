package room

import "w2g/internal/source"

type repository interface {
	GetGlobalRoom() (*Room, error)
	GetSourceById(string) (*source.Source, error)
	UpdateGlobalRoomSource(string) (string, error)
}

type service struct {
	repo repository
}

func NewService(r repository) *service {
	return &service{
		repo: r,
	}
}

func (s service) GetGlobalRoom() (*Room, error) {
	return s.repo.GetGlobalRoom()
}

func (s service) GetSourceById(id string) (*source.Source, error) {
	return s.repo.GetSourceById(id)
}

func (s service) UpdateGlobalRoomSource(srcID string) (string, error) {
	return s.repo.UpdateGlobalRoomSource(srcID)
}
