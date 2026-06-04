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
	Add(name, url string) (string, error)
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

	//nolint:errcheck
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
