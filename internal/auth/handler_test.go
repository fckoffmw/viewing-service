package auth

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHandler_Register_BadJSON(t *testing.T) {
	repo := &mockRepo{}
	store := &mockSessionStore{}
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
	repo := &mockRepo{}
	store := &mockSessionStore{}
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
	repo := &mockRepo{}
	store := &mockSessionStore{}
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
	store := &mockSessionStore{}
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