package room

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"

	httputils "w2g/internal/http"
	"w2g/internal/source"
)

var errHandler = errors.New("handler error")

type testService struct {
	resp        *GetResponse
	allResp     []GetResponse
	err         error
	deleteErr   error
}

func (m *testService) GetAllRooms() []GetResponse {
	return m.allResp
}

func (m *testService) SetCurrentSourceID(roomID, sourceID string) {}
func (m *testService) Create(req CreateRequest, ownerID string) (*CreateResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CreateResponse{
		ID:         "1",
		Name:       req.Name,
		InviteCode: "ABCD1234",
		InviteURL:  "/room/ABCD1234",
		OwnerID:    ownerID,
	}, nil
}

func (m *testService) GetByInviteCode(inviteCode string) (*GetResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func (m *testService) Delete(inviteCode string, userID string) error {
	if m.err != nil {
		return m.err
	}
	return m.deleteErr
}

func (m *testService) GetRoomByID(id string) (*Room, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &Room{ID: id}, nil
}

type testHub struct {
	membersOnline int
}

func (m *testHub) GetRoomState(roomID string) (string, string, bool, float64) {
	return "", "", false, 0
}

func (m *testHub) GetMembersOnline(roomID string) int {
	return m.membersOnline
}

func (m *testHub) BroadcastSourceChanged(roomID, sourceID, sourceURL string) {}

type testSourceStore struct {
	src *source.Source
	err error
}

func (m *testSourceStore) GetSourceById(id string) (*source.Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.src, nil
}

func TestHandlerCreateRoom(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantErr   error
	}{
		{
			name:       "success",
			body:      `{"name":"Movie Night"}`,
			wantStatus: http.StatusCreated,
			wantErr:   nil,
		},
		{
			name:       "invalid json",
			body:      `invalid`,
			wantStatus: http.StatusBadRequest,
			wantErr:   nil,
		},
		{
			name:       "empty name",
			body:      `{"name":""}`,
			wantStatus: http.StatusBadRequest,
			wantErr:   nil,
		},
		{
			name:       "service error",
			body:      `{"name":"Test"}`,
			wantStatus: http.StatusInternalServerError,
			wantErr:   errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{err: tt.wantErr}
			hub := &testHub{}
			srcStore := &testSourceStore{}
			h := NewHandler(svc, hub, srcStore, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/rooms", strings.NewReader(tt.body))
			ctx := context.WithValue(r.Context(), "user_id", "user1")
			r = r.WithContext(ctx)

			h.CreateRoom(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandlerGetRoom(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantErr    error
	}{
		{
			name:       "success",
			path:      "/api/rooms/ABCD1234",
			wantStatus: http.StatusOK,
			wantErr:   nil,
		},
		{
			name:       "not found",
			path:      "/api/rooms/NOTEXIST",
			wantStatus: http.StatusNotFound,
			wantErr:   errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{
				resp: &GetResponse{ID: "1", InviteCode: "ABCD1234"},
				err:  tt.wantErr,
			}
			hub := &testHub{}
			srcStore := &testSourceStore{}
			h := NewHandler(svc, hub, srcStore, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)

			h.GetRoom(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandlerDeleteRoom(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		userID     string
		wantStatus int
		wantErr    error
	}{
		{
			name:       "success as owner",
			path:      "/api/rooms/ABCD1234",
			userID:    "user1",
			wantStatus: http.StatusOK,
			wantErr:   nil,
		},
		{
			name:       "not owner",
			path:      "/api/rooms/ABCD1234",
			userID:    "user2",
			wantStatus: http.StatusForbidden,
			wantErr:   nil,
		},
		{
			name:       "not found",
			path:      "/api/rooms/NOTEXIST",
			userID:    "user1",
			wantStatus: http.StatusNotFound,
			wantErr:   errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{
				resp: &GetResponse{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"},
				err:  tt.wantErr,
			}
			hub := &testHub{}
			srcStore := &testSourceStore{}
			h := NewHandler(svc, hub, srcStore, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			ctx := context.WithValue(r.Context(), "user_id", tt.userID)
			r = r.WithContext(ctx)

			h.DeleteRoom(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandlerPatchRoomSource(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		userID     string
		wantStatus int
		wantErr    error
	}{
		{
			name:       "success",
			path:      "/api/rooms/ABCD1234/source",
			body:      `{"source_id":"1"}`,
			userID:    "user1",
			wantStatus: http.StatusOK,
			wantErr:   nil,
		},
		{
			name:       "not owner",
			path:      "/api/rooms/ABCD1234/source",
			body:      `{"source_id":"1"}`,
			userID:    "user2",
			wantStatus: http.StatusForbidden,
			wantErr:   nil,
		},
		{
			name:       "invalid json",
			path:      "/api/rooms/ABCD1234/source",
			body:      `invalid`,
			userID:    "user1",
			wantStatus: http.StatusBadRequest,
			wantErr:   nil,
		},
		{
			name:       "source not found",
			path:      "/api/rooms/ABCD1234/source",
			body:      `{"source_id":"999"}`,
			userID:    "user1",
			wantStatus: http.StatusNotFound,
			wantErr:   errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{
				resp: &GetResponse{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"},
				err:  tt.wantErr,
			}
			hub := &testHub{}
			srcStore := &testSourceStore{src: &source.Source{ID: "1"}, err: tt.wantErr}
			h := NewHandler(svc, hub, srcStore, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, tt.path, strings.NewReader(tt.body))
			ctx := context.WithValue(r.Context(), "user_id", tt.userID)
			r = r.WithContext(ctx)

			h.PatchRoomSource(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestExtractInviteCode(t *testing.T) {
	tests := []struct {
		name  string
		path string
		want string
	}{
		{
			name:  "simple path",
			path:  "/api/rooms/ABCD1234",
			want:  "ABCD1234",
		},
		{
			name:  "nested path",
			path:  "/api/rooms/ABCD1234/source",
			want:  "ABCD1234",
		},
		{
			name:  "no match",
			path:  "/api/other",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			got := httputils.ExtractInviteCode(r)
			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandlerGetAllRooms(t *testing.T) {
	tests := []struct {
		name       string
		rooms      []GetResponse
		wantStatus int
	}{
		{
			name:       "success with rooms",
			rooms:      []GetResponse{{ID: "1", Name: "Room1", InviteCode: "R1", OwnerID: "user1"}, {ID: "2", Name: "Room2", InviteCode: "R2", OwnerID: "user2"}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty rooms",
			rooms:      []GetResponse{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{allResp: tt.rooms}
			hub := &testHub{}
			srcStore := &testSourceStore{}
			h := NewHandler(svc, hub, srcStore, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)

			h.GetAllRooms(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}