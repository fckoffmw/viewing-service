package source

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"w2g/internal/response"
)

type Service interface {
	GetAllSources() ([]Source, error)
	AddSource(name, url string) (string, error)
}

type handler struct {
	service Service
	log     *slog.Logger
}

type AddSourceRequest struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type AddSourceResponse struct {
	ID string `json:"id"`
}

type SourceResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
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
		h.log.Error("when getting sources", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot get sources")
		return
	}

	h.log.Info("sources retrieved", "request_id", requestID, "count", len(sources))
	response.WriteOK(w, sources)
}

func (h handler) AddSource(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	var req AddSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)
		response.WriteBadRequest(w, "cannot read req body")
		return
	}
	r.Body.Close()

	if req.Name == "" || req.Url == "" {
		response.WriteBadRequest(w, "name and url are required")
		return
	}

	id, err := h.service.AddSource(req.Name, req.Url)
	if err != nil {
		h.log.Error("when adding source", "request_id", requestID, "err", err)
		response.WriteInternalError(w, "cannot add source")
		return
	}
	h.log.Info("successfuly added new source", "request_id", requestID, "id", id)

	response.WriteCreated(w, AddSourceResponse{ID: id})
}
