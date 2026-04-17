package source

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Service interface {
	GetAllSources() ([]Source, error)
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

func (h handler) GetAllSources(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	sources, err := h.service.GetAllSources()
	if err != nil {
		h.log.Error("when getting add sources", "request_id", requestID, "err", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.log.Info("sources retrieved", "request_id", requestID, "count", len(sources))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}
