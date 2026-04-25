package auth

import (
	"encoding/json"
	"net/http"
	"time"
)

type credentials struct {
	Username string `json:"username"`
	Pass     string `json:"password"`
}

type handler struct {
	service *service
}

func NewHandler(s *service) *handler {
	return &handler{
		service: s,
	}
}

func (h handler) Login(w http.ResponseWriter, r *http.Request) {
	// сверить логин пароль (хеш)
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
	}

	// создать сессию
}

func (h handler) Register(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if err := h.service.CheckUsernameAndPassword(creds.Username, creds.Pass); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	user, err := h.service.AddUser(creds.Username, creds.Pass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	sessionID := generateSessionID()

	if err := h.service.SaveSession(sessionID, creds.Username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	// Отправка ответа
	w.WriteHeader(http.StatusCreated)

}
