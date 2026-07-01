package source

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"
)

type mockService struct {
	sources     []Source
	addSourceID string
	err         error
}

func (m *mockService) GetAll() ([]Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}

func (m *mockService) Add(name, url string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.sources = append(m.sources, Source{
		ID:   m.addSourceID,
		Name: name,
		URL:  url,
	})
	return m.addSourceID, nil
}

func (m *mockService) Update(id, name, url string) error {
	if m.err != nil {
		return m.err
	}

	for i := range m.sources {
		if m.sources[i].ID == id {
			m.sources[i].Name = name
			m.sources[i].URL = url

			return nil
		}
	}

	return errRepo
}

func (m *mockService) Delete(id string) error {
	if m.err != nil {
		return m.err
	}

	for i := range m.sources {
		if m.sources[i].ID == id {
			m.sources = append(m.sources[:i], m.sources[i+1:]...)

			return nil
		}
	}

	return errRepo
}

func TestHandlerGetAll(t *testing.T) {
	tests := []struct {
		name       string
		sources    []Source
		wantLen    int
		wantFirst  string
		wantErr    error
		wantStatus int
	}{
		{
			name:       "empty",
			sources:    nil,
			wantLen:    0,
			wantFirst:  "",
			wantErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name: "success",
			sources: []Source{
				{ID: "1", Name: "Film", URL: "http://vk.com/1"},
				{ID: "2", Name: "Show", URL: "http://vk.com/2"},
			},
			wantLen:    2,
			wantFirst:  "Film",
			wantErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "service error",
			sources:    nil,
			wantLen:    0,
			wantFirst:  "",
			wantErr:    errRepo,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{sources: tt.sources, err: tt.wantErr}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/sources", nil)

			h.GetAll(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErr != nil {
				return
			}

			var got []Source
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0].Name != tt.wantFirst {
				t.Errorf("got[0].Name = %q, want %q", got[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestHandlerAdd(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		addSourceID  string
		wantID       string
		wantErr      error
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name:         "success",
			body:         `{"name":"New Film","url":"http://vk.com/3"}`,
			addSourceID:  "3",
			wantID:       "3",
			wantErr:      nil,
			wantStatus:   http.StatusCreated,
			wantErrorMsg: "",
		},
		{
			name:         "invalid json body",
			body:         `invalid`,
			addSourceID:  "",
			wantID:       "",
			wantErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "cannot read req body",
		},
		{
			name:         "service error",
			body:         `{"name":"New Film","url":"http://vk.com/3"}`,
			addSourceID:  "",
			wantID:       "",
			wantErr:      errRepo,
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "cannot add source",
		},
		{
			name:         "empty name",
			body:         `{"name":"","url":"http://vk.com/3"}`,
			addSourceID:  "",
			wantID:       "",
			wantErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "name and url are required",
		},
		{
			name:         "empty url",
			body:         `{"name":"New Film","url":""}`,
			addSourceID:  "",
			wantID:       "",
			wantErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "name and url are required",
		},
		{
			name:         "invalid fields",
			body:         `{"nameaaaa":"Film","urla":"http://vk.com/3"}`,
			addSourceID:  "",
			wantID:       "",
			wantErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "name and url are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{addSourceID: tt.addSourceID, err: tt.wantErr}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/sources", strings.NewReader(tt.body))

			h.Add(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErr != nil {
				return
			}

			if tt.wantStatus == http.StatusCreated {
				var resp AddResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if resp.ID != tt.wantID {
					t.Errorf("resp.ID = %q, want %q", resp.ID, tt.wantID)
				}
			}
		})
	}
}

func TestHandlerPatch(t *testing.T) {
	tests := []struct {
		name         string
		sourceID     string
		body         string
		initialSrcs  []Source
		wantStatus   int
		wantName     string
		wantErrorMsg string
	}{
		{
			name:        "success",
			sourceID:    "1",
			body:        `{"name":"Updated","url":"http://vk.com/new"}`,
			initialSrcs: []Source{{ID: "1", Name: "Old", URL: "http://vk.com/old"}},
			wantStatus:  http.StatusOK,
			wantName:    "Updated",
		},
		{
			name:         "invalid json",
			sourceID:     "1",
			body:         `invalid`,
			initialSrcs:  []Source{{ID: "1", Name: "Old", URL: "http://vk.com/old"}},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "cannot read req body",
		},
		{
			name:         "empty name",
			sourceID:     "1",
			body:         `{"name":"","url":"http://vk.com/new"}`,
			initialSrcs:  []Source{{ID: "1", Name: "Old", URL: "http://vk.com/old"}},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "name and url are required",
		},
		{
			name:         "empty url",
			sourceID:     "1",
			body:         `{"name":"Updated","url":""}`,
			initialSrcs:  []Source{{ID: "1", Name: "Old", URL: "http://vk.com/old"}},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "name and url are required",
		},
		{
			name:         "not found",
			sourceID:     "999",
			body:         `{"name":"Updated","url":"http://vk.com/new"}`,
			initialSrcs:  []Source{{ID: "1", Name: "Old", URL: "http://vk.com/old"}},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "cannot update source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{sources: tt.initialSrcs, err: errRepo}
			if tt.wantStatus == http.StatusOK || tt.wantStatus == http.StatusBadRequest {
				svc.err = nil
			}
			if tt.name == "not found" {
				svc.err = errRepo
			}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, "/api/sources/"+tt.sourceID, strings.NewReader(tt.body))
			r.SetPathValue("id", tt.sourceID)

			h.Patch(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp PatchResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if resp.ID != tt.sourceID {
					t.Errorf("resp.ID = %q, want %q", resp.ID, tt.sourceID)
				}
				if len(svc.sources) > 0 && svc.sources[0].Name != tt.wantName {
					t.Errorf("source.Name = %q, want %q", svc.sources[0].Name, tt.wantName)
				}
			}

			if tt.wantErrorMsg != "" {
				var errResp map[string]string
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if errResp["error"] != tt.wantErrorMsg {
					t.Errorf("error = %q, want %q", errResp["error"], tt.wantErrorMsg)
				}
			}
		})
	}
}

func TestHandlerDelete(t *testing.T) {
	tests := []struct {
		name         string
		sourceID     string
		initialSrcs  []Source
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name:        "success",
			sourceID:    "1",
			initialSrcs: []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			wantStatus:  http.StatusOK,
		},
		{
			name:         "not found",
			sourceID:     "999",
			initialSrcs:  []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "cannot delete source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{sources: tt.initialSrcs}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/api/sources/"+tt.sourceID, nil)
			r.SetPathValue("id", tt.sourceID)

			h.Delete(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp DeleteResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if len(svc.sources) != 0 {
					t.Errorf("sources len = %d, want 0", len(svc.sources))
				}
			}
		})
	}
}
