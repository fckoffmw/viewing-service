package room

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"w2g/internal/source"
)

type Service interface {
	GetGlobalRoom() (*Room, error)
	GetSourceById(string) (*source.Source, error)
}

type handler struct {
	service Service
	log     *slog.Logger
}

type RoomResponse struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	CurrentSource *source.Source `json:"current_source,omitempty"`
}

func NewHandler(s Service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func (h handler) GetGlobalRoom(w http.ResponseWriter, r *http.Request) {
	room, err := h.service.GetGlobalRoom()
	if err != nil {
		h.log.Error("when getting global room", "err", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := RoomResponse{
		ID:   room.ID,
		Name: "global",
	}

	if room.SourceID != "" {
		source, err := h.service.GetSourceById(room.SourceID)
		if err != nil {
			h.log.Error("when getting source", "err", err, "source_id", room.SourceID)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.CurrentSource = source
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h handler) PutGlobalRoomSource(w http.ResponseWriter, r *http.Request) {

}
