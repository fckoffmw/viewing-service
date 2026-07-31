package source

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"w2g/internal/http/response"
	"w2g/internal/utils/ctx"
)

type Service interface {
	GetAll() ([]Source, error)
	Add(name, strWithURL string) (string, error)
	Update(id, name, url string) error
	Delete(id string) error
}

type handler struct {
	service Service
	log     *slog.Logger
}

type AddRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AddResponse struct {
	ID string `json:"id"`
}

type PatchRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PatchResponse struct {
	ID string `json:"id"`
}

type DeleteResponse struct{}

func NewHandler(s Service, l *slog.Logger) *handler {
	return &handler{
		service: s,
		log:     l,
	}
}

func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())

	sources, err := h.service.GetAll()
	if err != nil {
		h.log.Error("failed to get sources", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot get sources")

		return
	}

	h.log.Debug("sources retrieved", "request_id", requestID, "count", len(sources))
	response.WriteOK(w, sources)
}

func (h *handler) Add(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())

	defer r.Body.Close()

	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "cannot read req body")

		return
	}

	if req.Name == "" || req.URL == "" {
		response.WriteBadRequest(w, "name and url are required")

		return
	}

	id, err := h.service.Add(req.Name, req.URL)
	if err != nil {
		h.log.Error("failed to add source", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot add source")

		return
	}

	h.log.Info("source added", "request_id", requestID, "id", id)
	response.WriteCreated(w, AddResponse{ID: id})
}

func (h *handler) Patch(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())
	id := r.PathValue("id")

	defer r.Body.Close()

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "cannot read req body")

		return
	}

	if req.Name == "" || req.URL == "" {
		response.WriteBadRequest(w, "name and url are required")

		return
	}

	if err := h.service.Update(id, req.Name, req.URL); err != nil {
		h.log.Error("failed to update source", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot update source")

		return
	}

	h.log.Info("source updated", "request_id", requestID, "id", id)
	response.WriteOK(w, PatchResponse{ID: id})
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := ctx.RequestIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.service.Delete(id); err != nil {
		h.log.Error("failed to delete source", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot delete source")

		return
	}

	h.log.Info("source deleted", "request_id", requestID, "id", id)
	response.WriteOK(w, DeleteResponse{})
}
