package auth

import (
	"errors"
	"fmt"
)

const (
	minPasswordLen = 8
)

type repository interface {
	GetUserByUsername(username string) (*User, error)
	AddUser(user User) (string, error)
}

type service struct {
	repo repository
}

func NewService(r repository) *service {
	return &service{
		repo: r,
	}
}

func (s service) Login(creds credentials) error {
	// достать пользоваля по имени
	user, err := s.repo.GetUserByUsername(creds.Username)
	if err != nil {

	}
	_ = user
	// сверить хеши
	// сгенерить сессию
	// вернуть сессию
	return nil
}

func (s *service) Register(username, password string) (*Session, error) {
	return &Session{}, nil
}

func (s *service) CheckUsernameAndPassword(username, password string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return err
	}

	if user == nil {
		return nil
	}

	if user.Username == username {
		return fmt.Errorf("user with username=%s has already exists", username)
	}

	if len(password) < minPasswordLen {
		return fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}

	return nil
}
