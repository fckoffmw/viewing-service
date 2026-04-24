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

func (m *mockService) GetAllSources() ([]Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}

func (m *mockService) AddSource(name, url string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.sources = append(m.sources, Source{
		ID:   m.addSourceID,
		Name: name,
		Url:  url,
	})
	return m.addSourceID, nil
}

func TestHandlerGetAllSources(t *testing.T) {
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
				{ID: "1", Name: "Film", Url: "http://vk.com/1"},
				{ID: "2", Name: "Show", Url: "http://vk.com/2"},
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

			h.GetAllSources(w, r)

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

func TestHandlerAddSource(t *testing.T) {
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
			wantStatus:   http.StatusOK,
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

			h.AddSource(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErr != nil {
				var resp SourceResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if resp.Error != tt.wantErrorMsg {
					t.Errorf("resp.Error = %q, want %q", resp.Error, tt.wantErrorMsg)
				}
				return
			}

			if tt.wantStatus == http.StatusOK {
				var resp SourceResponse
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
