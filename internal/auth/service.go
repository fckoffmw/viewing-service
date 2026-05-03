package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	apperrors "w2g/internal/errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionExpiry = 7 * 24 * time.Hour

	minUsernameLen = 3
	minPasswordLen = 4
	hashCost       = 12
)

func InvalidCredentials() *apperrors.Error { return apperrors.Unauthorized("invalid credentials") }
func EmptyUsername() *apperrors.Error      { return apperrors.BadRequest("username cannot be empty") }
func ShortUsername() *apperrors.Error {
	return apperrors.BadRequest("username must be at least 3 characters")
}
func ShortPassword() *apperrors.Error {
	return apperrors.BadRequest("password must be at least 4 characters")
}

type repository interface {
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id string) (*User, error)
	AddUser(user *User) (string, error)
}

type SessionStore interface {
	Set(*Session)
	Get(id string) (*Session, bool)
	Delete(id string)
}

type service struct {
	repo         repository
	sessionStore SessionStore
}

func NewService(r repository, s SessionStore) *service {
	return &service{
		repo:         r,
		sessionStore: s,
	}
}

func (s *service) Login(username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", InvalidCredentials()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", InvalidCredentials()
	}

	sessionID := generateSessionID()

	createdAt := time.Now()

	s.sessionStore.Set(&Session{
		SessionID:  sessionID,
		UserID:     user.ID,
		CreatedAt:  createdAt,
		LastSeenAt: createdAt,
		ExpiresAt:  createdAt.Add(SessionExpiry),
	})

	return sessionID, nil
}

func (s *service) Register(username, password string) (string, error) {
	if err := s.checkUsernameAndPasswordWhenRegister(username, password); err != nil {
		return "", err
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", err
	}

	createdAt := time.Now()

	userID, err := s.repo.AddUser(&User{
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
		ExpiresAt:  createdAt.Add(SessionExpiry),
	})

	return sessionID, nil
}

func (s *service) Logout(sessionID string) {
	s.sessionStore.Delete(sessionID)
}

func (s *service) GetUserBySession(sessionID string) (*User, error) {
	sess, ok := s.sessionStore.Get(sessionID)
	if !ok {
		return nil, InvalidCredentials()
	}

	user, err := s.repo.GetUserByID(sess.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, InvalidCredentials()
	}

	return user, nil
}

func (s *service) checkUsernameAndPasswordWhenRegister(username, password string) error {
	if username == "" {
		return EmptyUsername()
	}

	if len(username) < minUsernameLen {
		return ShortUsername()
	}

	if len(password) < minPasswordLen {
		return ShortPassword()
	}

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return apperrors.Internal(err.Error())
	}

	if user != nil {
		return apperrors.BadRequest("user already exists")
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
