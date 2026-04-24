package source

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Service interface {
	GetAllSources() ([]Source, error)
	AddSource(name, url string) (string, error)
}

type handler struct {
	service Service
	log     *slog.Logger
}

type Request struct {
	Name string `json:"name"`
	Url  string `json:"url"`
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

func writeError(w http.ResponseWriter, code int, resp SourceResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func (h handler) GetAllSources(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	sources, err := h.service.GetAllSources()
	if err != nil {
		h.log.Error("when getting sources", "request_id", requestID, "err", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.log.Info("sources retrieved", "request_id", requestID, "count", len(sources))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

func (h handler) AddSource(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value("request_id").(string)

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("when reading req body", "request_id", requestID, "err", err)

		writeError(w, http.StatusBadRequest, SourceResponse{
			Error: "cannot read req body",
		})
		return
	}
	defer r.Body.Close()

	if req.Name == "" || req.Url == "" {
		writeError(w, http.StatusBadRequest, SourceResponse{
			Error: "name and url are required",
		})
		return
	}

	id, err := h.service.AddSource(req.Name, req.Url)
	if err != nil {
		h.log.Error("when adding source", "request_id", requestID, "err", err)

		writeError(w, http.StatusInternalServerError, SourceResponse{
			Error: "cannot add source",
		})
		return
	}
	h.log.Info("successfuly added new source", "request_id", requestID, "id", id)

	resp := SourceResponse{
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
