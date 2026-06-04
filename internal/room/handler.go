package room

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"w2g/internal/utils/ctx"
	"w2g/internal/http/response"
	"w2g/internal/source"
)

type Service interface {
	Create(req CreateRequest, ownerID string) (*CreateResponse, error)
	GetByInviteCode(inviteCode string) (*GetResponse, error)
	Delete(inviteCode string, userID string) error
	GetAll() []GetResponse
	PatchSource(inviteCode, ownerID, sourceID string) error
}

type CreateRequest struct {
	Name string `json:"name"`
}

type CreateResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
	InviteURL  string `json:"invite_url"`
	OwnerID    string `json:"owner_id"`
	CreatedAt  string `json:"created_at"`
}

type GetResponse struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	InviteCode    string         `json:"invite_code"`
	OwnerID       string         `json:"owner_id"`
	MembersOnline int            `json:"members_online"`
	CurrentSource *source.Source `json:"current_source,omitempty"`
	CreatedAt     string         `json:"created_at"`
}

type PatchRequest struct {
	SourceID string `json:"source_id"`
}

type PatchResponse struct {
	SourceID string `json:"source_id"`
}

type handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(svc Service, l *slog.Logger) *handler {
	return &handler{
		service: svc,
		log:     l,
	}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())
	userID := ctx.UserIDFromContext(r.Context())

	//nolint:errcheck
	//nolint:errcheck
	defer r.Body.Close()

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "invalid request body")

		return
	}

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

		h.log.Error("failed to create room", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot create room")

		return
	}

	h.log.Info("room created", "request_id", requestID, "room_id", resp.ID, "invite_code", resp.InviteCode)

	response.WriteCreated(w, CreateResponse{
		ID:         resp.ID,
		Name:       resp.Name,
		InviteCode: resp.InviteCode,
		InviteURL:  resp.InviteURL,
		OwnerID:    resp.OwnerID,
		CreatedAt:  resp.CreatedAt,
	})
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())
	inviteCode := r.PathValue("invite_code")

	room, err := h.service.GetByInviteCode(inviteCode)
	if err != nil {
		h.log.Warn("room not found", "request_id", requestID, "invite_code", inviteCode)
		response.WriteNotFound(w, "room not found")

		return
	}

	response.WriteOK(w, GetResponse{
		ID:            room.ID,
		Name:          room.Name,
		InviteCode:    room.InviteCode,
		OwnerID:       room.OwnerID,
		MembersOnline: room.MembersOnline,
		CurrentSource: room.CurrentSource,
		CreatedAt:     room.CreatedAt,
	})
}

func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {
	rooms := h.service.GetAll()
	response.WriteOK(w, rooms)
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())
	userID := ctx.UserIDFromContext(r.Context())
	inviteCode := r.PathValue("invite_code")

	room, err := h.service.GetByInviteCode(inviteCode)
	if err != nil {
		h.log.Warn("room not found for delete", "request_id", requestID, "invite_code", inviteCode)
		response.WriteNotFound(w, "room not found")

		return
	}

	if room.OwnerID != userID {
		response.WriteForbidden(w, "not owner")

		return
	}

	if err := h.service.Delete(inviteCode, userID); err != nil {
		h.log.Error("failed to delete room", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot delete room")

		return
	}

	h.log.Info("room deleted", "request_id", requestID, "room_id", room.ID)

	response.WriteOK(w, map[string]string{"status": "deleted"})
}

func (h *handler) PatchSource(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())
	userID := ctx.UserIDFromContext(r.Context())
	inviteCode := r.PathValue("invite_code")

	//nolint:errcheck
	defer r.Body.Close()

	room, err := h.service.GetByInviteCode(inviteCode)
	if err != nil {
		h.log.Warn("room not found for patch", "request_id", requestID, "invite_code", inviteCode)
		response.WriteNotFound(w, "room not found")

		return
	}

	if room.OwnerID != userID {
		response.WriteForbidden(w, "not owner")

		return
	}

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "invalid request body")

		return
	}

	if err := h.service.PatchSource(inviteCode, userID, req.SourceID); err != nil {
		if errors.Is(err, ErrSourceNotFound) {
			response.WriteBadRequest(w, "source not found")

			return
		}
		h.log.Error("failed to patch source", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot change source")

		return
	}

	h.log.Info("source changed", "request_id", requestID, "room_id", room.ID, "source_id", req.SourceID)

	response.WriteOK(w, PatchResponse(req))
}
