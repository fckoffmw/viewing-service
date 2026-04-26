package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	apperrors "w2g/internal/errors"
)

type Service interface {
	Login(username, password string) (string, error)
	Register(username, password string) (string, error)
	Logout(sessionID string)
	GetUserBySession(sessionID string) (*User, error)
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type handler struct {
	service Service
	log     *slog.Logger
}

type AuthResponse struct {
	Error string `json:"error,omitempty"`
}

func NewHandler(s Service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func writeError(w http.ResponseWriter, code int, resp AuthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func (h handler) Login(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)

		writeError(w, http.StatusBadRequest, AuthResponse{Error: "cannot read req body"})
		return
	}
	defer r.Body.Close()

	sessionID, err := h.service.Login(creds.Username, creds.Password)
	if err != nil {
		h.log.Error("when login", "request_id", requestID, "err", err)

		var appErr *apperrors.Error
		if errors.As(err, &appErr) {
			writeError(w, appErr.Code, AuthResponse{Error: appErr.Message})
			return
		}

		writeError(w, http.StatusInternalServerError, AuthResponse{Error: "internal server error"})
		return
	}

	h.log.Info("successful login", "request_id", requestID, "username", creds.Username)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{})
}

func (h handler) Register(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)

		writeError(w, http.StatusBadRequest, AuthResponse{Error: "cannot read req body"})
		return
	}
	defer r.Body.Close()

	sessionID, err := h.service.Register(creds.Username, creds.Password)
	if err != nil {
		h.log.Error("when register", "request_id", requestID, "err", err)

		var appErr *apperrors.Error
		if errors.As(err, &appErr) {
			writeError(w, appErr.Code, AuthResponse{Error: appErr.Message})
			return
		}

		writeError(w, http.StatusInternalServerError, AuthResponse{Error: "internal server error"})
		return
	}

	h.log.Info("successful register", "request_id", requestID, "username", creds.Username)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{})
}

func (h handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.service.Logout(cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}

type MeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h handler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		writeError(w, http.StatusUnauthorized, AuthResponse{Error: "session not found"})
		return
	}

	user, err := h.service.GetUserBySession(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, AuthResponse{Error: "session not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MeResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}