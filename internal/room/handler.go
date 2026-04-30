package room

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"w2g/internal/response"
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
	Error         string         `json:"error,omitempty"`
}

func NewHandler(s Service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func (h handler) GetGlobalRoom(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	room, err := h.service.GetGlobalRoom()
	if err != nil {
		h.log.Error("when getting global room", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot get room")
		return
	}

	resp := RoomResponse{
		ID:   room.ID,
		Name: "global",
	}

	if room.SourceID != "" {
		src, err := h.service.GetSourceById(room.SourceID)
		if err != nil {
			h.log.Error("when getting source", "request_id", requestID, "err", err, "source_id", room.SourceID)
			response.WriteInternalError(w, "cannot get source")
			return
		}
		resp.CurrentSource = src
	}

	response.WriteOK(w, resp)
}

func (h handler) PatchGlobalRoomSource(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "cannot read req body")
		return
	}
	r.Body.Close()

	if _, err := h.service.GetSourceById(req.SourceID); err != nil {
		h.log.Error("when getting source with id", "request_id", requestID, "id", req.SourceID, "err", err)
		response.WriteBadRequest(w, "source not found")
		return
	}

	id, err := h.service.UpdateGlobalRoomSource(req.SourceID)
	if err != nil {
		h.log.Error("when updating global room source", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot update room")
		return
	}
	h.log.Info("successfuly updated global room source", "request_id", requestID, "id", id)

	response.WriteOK(w, RoomResponse{ID: id})
}
