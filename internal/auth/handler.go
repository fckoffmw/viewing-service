package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type handler struct {
	service *service
	log     *slog.Logger
}

type AuthResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewHandler(s *service, l *slog.Logger) *handler {
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

		writeError(w, http.StatusBadRequest, AuthResponse{
			Error: "cannot read req body",
		})
		return
	}
	defer r.Body.Close()

	sessionID, err := h.service.Login(creds.Username, creds.Password)
	if err != nil {
		h.log.Error("when login", "request_id", requestID, "err", err)

		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, AuthResponse{
				Error: "invalid credentials",
			})
			return
		}

		writeError(w, http.StatusInternalServerError, AuthResponse{
			Error: "internal server error",
		})
		return
	}

	h.log.Info("successful login", "request_id", requestID, "username", creds.Username)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   false, // ! TODO
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

func (h handler) Register(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)

		writeError(w, http.StatusBadRequest, AuthResponse{
			Error: "cannot read req body",
		})
		return
	}
	defer r.Body.Close()

	sessionID, err := h.service.Register(creds.Username, creds.Password)
	if err != nil {
		h.log.Error("when register", "request_id", requestID, "err", err)

		if errors.Is(err, ErrUserAlreadyExists) || errors.Is(err, ErrShortPassword) || errors.Is(err, ErrEmptyUsername) {
			writeError(w, http.StatusBadRequest, AuthResponse{
				Error: err.Error(),
			})
			return
		}

		writeError(w, http.StatusInternalServerError, AuthResponse{
			Error: "internal server error",
		})
		return
	}

	h.log.Info("successful register", "request_id", requestID, "username", creds.Username)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusCreated)
}
