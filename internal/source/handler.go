package source

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type handler struct {
	service *service
	log     *slog.Logger
}

func NewHandler(s *service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func (h handler) GetAllSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.service.GetAllSources()
	if err != nil {
		h.log.Error("when getting add sources", "err", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}
