package auth

import (
	"encoding/json"
	"net/http"
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
