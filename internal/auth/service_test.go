package auth

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
	apperrors "w2g/internal/errors"
)

type mockRepo struct {
	user *User
	err  error
}

func (m *mockRepo) GetUserByUsername(username string) (*User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockRepo) GetUserByID(id string) (*User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockRepo) AddUser(user User) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "new-user-id", nil
}

type mockSessionStore struct {
	sessions map[string]*Session
}

func (m *mockSessionStore) Set(sess *Session) {
	if m.sessions == nil {
		m.sessions = make(map[string]*Session)
	}
	m.sessions[sess.SessionID] = sess
}

func (m *mockSessionStore) Get(id string) (*Session, bool) {
	if m.sessions == nil {
		return nil, false
	}
	sess, ok := m.sessions[id]
	return sess, ok
}

func (m *mockSessionStore) Delete(id string) {
	if m.sessions != nil {
		delete(m.sessions, id)
	}
}

func TestService_Register(t *testing.T) {
	t.Run("empty username", func(t *testing.T) {
		repo := &mockRepo{}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Register("", "password123")

		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != 400 {
			t.Errorf("expected 400 error, got %v", err)
		}
	})

	t.Run("short password", func(t *testing.T) {
		repo := &mockRepo{}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Register("user", "abc")

		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != 400 {
			t.Errorf("expected 400 error, got %v", err)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		repo := &mockRepo{user: &User{Username: "existing"}}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Register("existing", "password123")

		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != 400 {
			t.Errorf("expected 400 error, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		sessionID, err := svc.Register("newuser", "password123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sessionID == "" {
			t.Error("expected non-empty session ID")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockRepo{err: errors.New("repo error")}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Register("user", "password123")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Error("expected non-empty session ID")
	}
	if len(id1) != 44 {
		t.Errorf("expected length 44, got %d", len(id1))
	}
	if id1 == id2 {
		t.Error("expected unique session IDs")
	}
}

func TestService_Login(t *testing.T) {
	t.Run("user not found", func(t *testing.T) {
		repo := &mockRepo{user: nil}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Login("nonexistent", "password123")

		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != 401 {
			t.Errorf("expected 401 error, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		repo := &mockRepo{user: &User{ID: "1", Username: "user", PasswordHash: "$2a$12$hash"}}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Login("user", "wrongpassword")

		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != 401 {
			t.Errorf("expected 401 error, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 12)
		repo := &mockRepo{user: &User{ID: "1", Username: "user", PasswordHash: string(hash)}}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		sessionID, err := svc.Login("user", "correctpassword")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sessionID == "" {
			t.Error("expected non-empty session ID")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockRepo{err: errors.New("repo error")}
		store := &mockSessionStore{}
		svc := NewService(repo, store)

		_, err := svc.Login("user", "password123")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}