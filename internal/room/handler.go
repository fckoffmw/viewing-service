package room

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"w2g/internal/source"
)

type Service interface {
	GetGlobalRoom() (*Room, error)
	GetSourceById(string) (*source.Source, error)
	UpdateGlobalRoomSource(string) (string, error)
}

type handler struct {
	service Service
	log     *slog.Logger
}

type Request struct {
	SourceID string `json:"source_id"`
}

type RoomResponse struct {
	ID            string         `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	CurrentSource *source.Source `json:"current_source,omitempty"`
	Message       string         `json:"message,omitempty"`
}

func NewHandler(s Service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func writeError(w http.ResponseWriter, code int, resp RoomResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
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

func (h handler) PatchGlobalRoomSource(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 || parts[1] != "api" || parts[2] != "room" || parts[3] != "source" {
		h.log.Error("invalid req path", "path", r.URL.Path)

		writeError(w, http.StatusBadRequest, RoomResponse{
			Message: "invalid req path",
		})
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("when reading req body", "err", err)

		writeError(w, http.StatusBadRequest, RoomResponse{
			Message: "cannot read req body",
		})
		return
	}
	defer r.Body.Close()

	if _, err := h.service.GetSourceById(req.SourceID); err != nil {
		h.log.Error("when getting source with id", "id", req.SourceID, "err", err)

		writeError(w, http.StatusBadRequest, RoomResponse{
			Message: "cannot get source with id " + req.SourceID,
		})
		return
	}

	id, err := h.service.UpdateGlobalRoomSource(req.SourceID)
	if err != nil {
		h.log.Error("when updating global room source", "err", err)

		writeError(w, http.StatusInternalServerError, RoomResponse{
			Message: "cannot update global room source",
		})
		return
	}
	h.log.Info("successfuly updated global room source", "id", id)

	resp := RoomResponse{
		ID:      id,
		Message: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
