package auth

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type testSessionStore struct {
	sessions map[string]*Session
}

type testRepo struct {
	err   error
	users []User
}

func (r *testRepo) GetUserByUsername(username string) (*User, error) {
	if r.err != nil {
		return nil, r.err
	}
	for _, u := range r.users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, nil
}

func (r *testRepo) GetUserByID(id string) (*User, error) {
	if r.err != nil {
		return nil, r.err
	}
	for _, u := range r.users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

func (r *testRepo) AddUser(user *User) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	r.users = append(r.users, *user)
	return "new-user-id", nil
}

func (m *testSessionStore) Set(sess *Session) {
	if m.sessions == nil {
		m.sessions = make(map[string]*Session)
	}
	m.sessions[sess.SessionID] = sess
}

func (m *testSessionStore) Get(id string) (*Session, bool) {
	if m.sessions == nil {
		return nil, false
	}
	sess, ok := m.sessions[id]
	if !ok || sess.ExpiresAt.Before(time.Now()) {
		return nil, false
	}
	return sess, ok
}

func (m *testSessionStore) Delete(id string) {
	if m.sessions != nil {
		delete(m.sessions, id)
	}
}

func TestHandler_Register_BadJSON(t *testing.T) {
	repo := &testRepo{}
	store := &testSessionStore{}
	svc := NewService(repo, store)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandler_Register_Success(t *testing.T) {
	repo := &testRepo{}
	store := &testSessionStore{}
	svc := NewService(repo, store)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandler(svc, logger)

	body := `{"username":"newuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestHandler_Login_BadJSON(t *testing.T) {
	repo := &testRepo{}
	store := &testSessionStore{}
	svc := NewService(repo, store)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	repo := &mockRepo{user: &User{ID: "1", Username: "user", PasswordHash: string(hash)}}
	store := &testSessionStore{}
	svc := NewService(repo, store)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandler(svc, logger)

	body := `{"username":"user","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_Login_Unauthorized(t *testing.T) {
	repo := &testRepo{}
	store := &testSessionStore{}
	svc := NewService(repo, store)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandler(svc, logger)

	body := `{"username":"user","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}