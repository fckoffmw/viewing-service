package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"w2g/internal/apperrors"
	"w2g/internal/utils/ctx"
	"w2g/internal/http/response"
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

func NewHandler(s Service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())

	//nolint:errcheck
	defer r.Body.Close()

	var creds credentials
	if err := decodeJSON(r, &creds); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "cannot read req body")
		return
	}

	sessionID, err := h.service.Login(creds.Username, creds.Password)
	if err != nil {
		h.log.Error("when login", "request_id", requestID, "err", err)

		var appErr *apperrors.Error
		if errors.As(err, &appErr) {
			response.WriteError(w, appErr.Code, appErr.Message)
			return
		}

		response.WriteInternalError(w, "internal server error")
		return
	}

	h.log.Info("successful login", "request_id", requestID, "username", creds.Username)

	setSessionCookie(w, sessionID)
	response.WriteOK(w, nil)
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())

	//nolint:errcheck
	defer r.Body.Close()

	var creds credentials
	if err := decodeJSON(r, &creds); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "cannot read req body")
		return
	}

	sessionID, err := h.service.Register(creds.Username, creds.Password)
	if err != nil {
		h.log.Error("when register", "request_id", requestID, "err", err)

		var appErr *apperrors.Error
		if errors.As(err, &appErr) {
			response.WriteError(w, appErr.Code, appErr.Message)
			return
		}

		response.WriteInternalError(w, "internal server error")
		return
	}

	h.log.Info("successful register", "request_id", requestID, "username", creds.Username)

	setSessionCookie(w, sessionID)
	response.WriteCreated(w, nil)
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.service.Logout(cookie.Value)

	clearSessionCookie(w)
	w.WriteHeader(http.StatusOK)
}

func (h *handler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		response.WriteUnauthorized(w, "session not found")
		return
	}

	user, err := h.service.GetUserBySession(cookie.Value)
	if err != nil {
		response.WriteUnauthorized(w, "session not found")
		return
	}

	response.WriteOK(w, MeResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}

type MeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func setSessionCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    value,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
