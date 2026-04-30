package room

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"

	"w2g/internal/source"
)

var (
	errTest   = errors.New("test error")
	errUpdate = errors.New("repo error")
)

type mockService struct {
	room      *Room
	src       *source.Source
	srcID     string
	err       error
	updateErr error
}

func (m *mockService) GetGlobalRoom() (*Room, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.room, nil
}

func (m *mockService) GetSourceById(id string) (*source.Source, error) {
	if m.err != nil && m.err != errUpdate {
		return nil, m.err
	}
	if m.src != nil && m.src.ID == id {
		return m.src, nil
	}
	return nil, errors.New("not found")
}

func (m *mockService) UpdateGlobalRoomSource(srcID string) (string, error) {
	if m.updateErr != nil {
		return "", m.updateErr
	}
	m.srcID = srcID
	return srcID, nil
}

func TestHandlerGetGlobalRoom(t *testing.T) {
	tests := []struct {
		name        string
		room        *Room
		src         *source.Source
		wantID      string
		wantSrcName string
		wantStatus  int
		wantErr     error
	}{
		{
			name:        "success without source",
			room:        &Room{ID: "1", SourceID: ""},
			src:         nil,
			wantID:      "1",
			wantSrcName: "",
			wantStatus:  http.StatusOK,
			wantErr:     nil,
		},
		{
			name:        "success with source",
			room:        &Room{ID: "1", SourceID: "5"},
			src:         &source.Source{ID: "5", Name: "Film", Url: "http://vk.com/1"},
			wantID:      "1",
			wantSrcName: "Film",
			wantStatus:  http.StatusOK,
			wantErr:     nil,
		},
		{
			name:        "service error",
			room:        nil,
			src:         nil,
			wantID:      "",
			wantSrcName: "",
			wantStatus:  http.StatusInternalServerError,
			wantErr:     errTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{room: tt.room, src: tt.src, err: tt.wantErr}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/room", nil)

			h.GetGlobalRoom(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErr != nil {
				return
			}

			var resp RoomResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if resp.ID != tt.wantID {
				t.Errorf("resp.ID = %q, want %q", resp.ID, tt.wantID)
			}
			if tt.wantSrcName != "" && resp.CurrentSource.Name != tt.wantSrcName {
				t.Errorf("resp.CurrentSource.Name = %q, want %q", resp.CurrentSource.Name, tt.wantSrcName)
			}
		})
	}
}

func TestHandlerPatchGlobalRoomSource(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		room       *Room
		src        *source.Source
		wantID     string
		wantError  string
		wantStatus int
		wantErr    error
		updateErr  error
	}{
		{
			name:       "success",
			path:       "/api/room/source",
			body:       `{"source_id":"5"}`,
			room:       &Room{ID: "1", SourceID: "1"},
			src:        &source.Source{ID: "5", Name: "Film", Url: "http://vk.com/1"},
			wantID:     "5",
			wantError:  "",
			wantStatus: http.StatusOK,
			wantErr:    nil,
			updateErr:  nil,
		},
		{
			name:       "invalid json body",
			path:       "/api/room/source",
			body:       `invalid`,
			room:       nil,
			src:        nil,
			wantID:     "",
			wantError:  "cannot read req body",
			wantStatus: http.StatusBadRequest,
			wantErr:    nil,
			updateErr:  nil,
		},
		{
			name:       "source not found",
			path:       "/api/room/source",
			body:       `{"source_id":"999"}`,
			room:       nil,
			src:        nil,
			wantID:     "",
			wantError:  "source not found",
			wantStatus: http.StatusBadRequest,
			wantErr:    nil,
			updateErr:  nil,
		},
		{
			name:       "service error",
			path:       "/api/room/source",
			body:       `{"source_id":"5"}`,
			room:       nil,
			src:        &source.Source{ID: "5", Name: "Film", Url: "http://vk.com/1"},
			wantID:     "",
			wantError:  "cannot update room",
			wantStatus: http.StatusInternalServerError,
			wantErr:    nil,
			updateErr:  errUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{room: tt.room, src: tt.src, err: tt.wantErr, updateErr: tt.updateErr}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, tt.path, strings.NewReader(tt.body))

			h.PatchGlobalRoomSource(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp RoomResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if resp.ID != tt.wantID {
				t.Errorf("resp.ID = %q, want %q", resp.ID, tt.wantID)
			}
			if resp.Error != tt.wantError {
				t.Errorf("resp.Error = %q, want %q", resp.Error, tt.wantError)
			}
		})
	}
}
