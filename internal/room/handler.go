package room

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"w2g/internal/response"
	"w2g/internal/source"
)

type HubManager interface {
	GetRoomState(roomID string) (sourceID string, sourceURL string, playing bool, position float64)
	GetMembersOnline(roomID string) int
	BroadcastSourceChanged(roomID, sourceID, sourceURL string)
}

type SourceStore interface {
	GetSourceById(id string) (*source.Source, error)
}

type ServiceHandler interface {
	Create(req CreateRequest, ownerID string) (*CreateResponse, error)
	GetByInviteCode(inviteCode string) (*GetResponse, error)
	Delete(inviteCode string, userID string) error
	GetRoomByID(id string) (*Room, error)
}

type handler struct {
	service     ServiceHandler
	hub         HubManager
	sourceStore SourceStore
	log        *slog.Logger
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type CreateRoomResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
	InviteURL  string `json:"invite_url"`
	OwnerID    string `json:"owner_id"`
	CreatedAt  string `json:"created_at"`
}

type GetRoomResponse struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	InviteCode     string         `json:"invite_code"`
	OwnerID      string         `json:"owner_id"`
	MembersOnline int            `json:"members_online"`
	CurrentSource *source.Source `json:"current_source,omitempty"`
	CreatedAt    string        `json:"created_at"`
}

type PatchSourceRequest struct {
	SourceID string `json:"source_id"`
}

type PatchSourceResponse struct {
	SourceID string `json:"source_id"`
}

func NewHandler(svc ServiceHandler, hub HubManager, srcStore SourceStore, l *slog.Logger) *handler {
	return &handler{
		service:     svc,
		hub:         hub,
		sourceStore: srcStore,
		log:         l,
	}
}

func (h handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)
	userID, _ := r.Context().Value("user_id").(string)

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "invalid request body")
		return
	}
	r.Body.Close()

	if req.Name == "" {
		response.WriteBadRequest(w, "name is required")
		return
	}

	resp, err := h.service.Create(CreateRequest{Name: req.Name}, userID)
	if err != nil {
		if errors.Is(err, ErrMaxRoomsReached) {
			response.WriteBadRequest(w, "max rooms reached")
			return
		}
		h.log.Error("when creating room", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot create room")
		return
	}

	h.log.Info("room created", "request_id", requestID, "room_id", resp.ID, "invite_code", resp.InviteCode)

	response.WriteCreated(w, CreateRoomResponse{
		ID:         resp.ID,
		Name:       resp.Name,
		InviteCode: resp.InviteCode,
		InviteURL:  resp.InviteURL,
		OwnerID:    resp.OwnerID,
		CreatedAt:  resp.CreatedAt,
	})
}

func (h handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)
	inviteCode := extractInviteCode(r)

	room, err := h.service.GetByInviteCode(inviteCode)
	if err != nil {
		h.log.Error("when getting room", "request_id", requestID, "invite_code", inviteCode, "err", err)
		response.WriteNotFound(w, "room not found")
		return
	}

	membersOnline := h.hub.GetMembersOnline(room.ID)

	response.WriteOK(w, GetRoomResponse{
		ID:            room.ID,
		Name:          room.Name,
		InviteCode:    room.InviteCode,
		OwnerID:       room.OwnerID,
		MembersOnline: membersOnline,
		CreatedAt:     room.CreatedAt,
	})
}

func (h handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	inviteCode := extractInviteCode(r)

	room, err := h.service.GetByInviteCode(inviteCode)
	if err != nil {
		h.log.Error("when getting room", "request_id", requestID, "invite_code", inviteCode, "err", err)
		response.WriteNotFound(w, "room not found")
		return
	}

	if room.OwnerID != userID {
		response.WriteForbidden(w, "not owner")
		return
	}

	if err := h.service.Delete(inviteCode, userID); err != nil {
		h.log.Error("when deleting room", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot delete room")
		return
	}

	h.log.Info("room deleted", "request_id", requestID, "room_id", room.ID)

	response.WriteOK(w, map[string]string{"status": "deleted"})
}

func (h handler) PatchRoomSource(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	inviteCode := extractInviteCode(r)

	room, err := h.service.GetByInviteCode(inviteCode)
	if err != nil {
		h.log.Error("when getting room", "request_id", requestID, "invite_code", inviteCode, "err", err)
		response.WriteNotFound(w, "room not found")
		return
	}

	if room.OwnerID != userID {
		response.WriteForbidden(w, "not owner")
		return
	}

	var req PatchSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "invalid request body")
		return
	}
	r.Body.Close()

	src, err := h.sourceStore.GetSourceById(req.SourceID)
	if err != nil {
		response.WriteBadRequest(w, "source not found")
		return
	}

	h.hub.BroadcastSourceChanged(room.ID, src.ID, src.Url)

	h.log.Info("source changed", "request_id", requestID, "room_id", room.ID, "source_id", req.SourceID)

	response.WriteOK(w, PatchSourceResponse{SourceID: req.SourceID})
}

func extractInviteCode(r *http.Request) string {
	// Try to get from URL path value first (works for patterns like /api/rooms/{invite_code})
	if inviteCode := r.PathValue("invite_code"); inviteCode != "" {
		return inviteCode
	}
	
	// Fallback: extract from path
	// For /api/rooms/INVITE_CODE/source -> extract INVITE_CODE (2nd from end)
	// For /api/rooms/INVITE_CODE -> extract INVITE_CODE (last)
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rooms/"), "/")
	if len(parts) >= 1 && parts[0] != "" {
		// parts[0] = invite_code, or invite_code/source
		inviteCode := strings.Split(parts[0], "/")[0]
		return inviteCode
	}
	return ""
}