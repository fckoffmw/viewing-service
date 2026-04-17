package source

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
)

type mockService struct {
	sources []Source
	err     error
}

func (m *mockService) GetAllSources() ([]Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
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
