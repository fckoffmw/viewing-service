package frame

import (
	"encoding/json"
	"net/http"
)

type handler struct {
	service *service
}

func NewHandler(s *service) *handler {
	return &handler{
		service: s,
	}
}

func (h handler) GetAllFrames(w http.ResponseWriter, r *http.Request) {
	frames, err := h.service.GetAllFrames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(frames)
}
