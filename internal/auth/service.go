package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("password must be at least 4 characters")
	ErrEmptyUsername      = errors.New("username cannot be empty")
)

const (
	minPasswordLen = 4
	hashCost       = 12
	sessionExpiry  = 7 * 24 * time.Hour
)

type repository interface {
	GetUserByUsername(username string) (*User, error)
	AddUser(user User) (string, error)
}

type sessionStore interface {
	Set(*Session)
}

type service struct {
	repo         repository
	sessionStore sessionStore
}

func NewService(r repository, s sessionStore) *service {
	return &service{
		repo:         r,
		sessionStore: s,
	}
}

func (s *service) Register(username, password string) (string, error) {
	if err := s.checkUsernameAndPassword(username, password); err != nil {
		return "", err
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", err
	}

	createdAt := time.Now()

	userID, err := s.repo.AddUser(User{
		Username:     username,
		PasswordHash: string(passHash),
		CreatedAt:    createdAt,
	})
	if err != nil {
		return "", err
	}

	sessionID := generateSessionID()

	s.sessionStore.Set(&Session{
		SessionID:  sessionID,
		UserID:     userID,
		CreatedAt:  createdAt,
		LastSeenAt: createdAt,
		ExpiresAt:  createdAt.Add(sessionExpiry),
	})

	return sessionID, nil
}

func (s *service) Login(username, password string) (string, error) {
	return "", nil
}

func (s *service) checkUsernameAndPassword(username, password string) error {
	if username == "" {
		return ErrEmptyUsername
	}

	if len(password) < minPasswordLen {
		return ErrInvalidPassword
	}

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return err
	}

	if user != nil {
		return ErrUserAlreadyExists
	}

	return nil
}

func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}
