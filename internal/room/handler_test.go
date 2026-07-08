package room

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"

	"w2g/internal/utils/ctx"
)

var errHandler = errors.New("handler error")

type testService struct {
	resp      *GetResponse
	allResp   []GetResponse
	err       error
	deleteErr error
	patchErr  error
}

func (m *testService) GetAll() []GetResponse {
	return m.allResp
}

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

func (m *testService) PatchSource(inviteCode, ownerID, sourceID string) error {
	return m.patchErr
}

func TestHandlerCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantErr    error
	}{
		{
			name:       "success",
			body:       `{"name":"Movie Night"}`,
			wantStatus: http.StatusCreated,
			wantErr:    nil,
		},
		{
			name:       "invalid json",
			body:       `invalid`,
			wantStatus: http.StatusBadRequest,
			wantErr:    nil,
		},
		{
			name:       "empty name",
			body:       `{"name":""}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    nil,
		},
		{
			name:       "service error",
			body:       `{"name":"Test"}`,
			wantStatus: http.StatusInternalServerError,
			wantErr:    errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{err: tt.wantErr}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/rooms", strings.NewReader(tt.body))
			r = r.WithContext(ctx.WithUserID(r.Context(), "user1"))

			h.Create(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandlerGet(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantErr    error
	}{
		{
			name:       "success",
			path:       "/api/rooms/ABCD1234",
			wantStatus: http.StatusOK,
			wantErr:    nil,
		},
		{
			name:       "not found",
			path:       "/api/rooms/NOTEXIST",
			wantStatus: http.StatusNotFound,
			wantErr:    errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{
				resp: &GetResponse{ID: "1", InviteCode: "ABCD1234"},
				err:  tt.wantErr,
			}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)

			h.Get(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandlerDelete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		userID     string
		wantStatus int
		wantErr    error
	}{
		{
			name:       "success as owner",
			path:       "/api/rooms/ABCD1234",
			userID:     "user1",
			wantStatus: http.StatusOK,
			wantErr:    nil,
		},
		{
			name:       "not owner",
			path:       "/api/rooms/ABCD1234",
			userID:     "user2",
			wantStatus: http.StatusForbidden,
			wantErr:    nil,
		},
		{
			name:       "not found",
			path:       "/api/rooms/NOTEXIST",
			userID:     "user1",
			wantStatus: http.StatusNotFound,
			wantErr:    errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{
				resp: &GetResponse{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"},
				err:  tt.wantErr,
			}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			r = r.WithContext(ctx.WithUserID(r.Context(), tt.userID))

			h.Delete(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandlerPatchSource(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		userID     string
		patchErr   error
		wantStatus int
	}{
		{
			name:       "success",
			path:       "/api/rooms/ABCD1234/source",
			body:       `{"source_id":"1"}`,
			userID:     "user1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "not owner",
			path:       "/api/rooms/ABCD1234/source",
			body:       `{"source_id":"1"}`,
			userID:     "user2",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid json",
			path:       "/api/rooms/ABCD1234/source",
			body:       `invalid`,
			userID:     "user1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "source not found",
			path:       "/api/rooms/ABCD1234/source",
			body:       `{"source_id":"999"}`,
			userID:     "user1",
			patchErr:   ErrSourceNotFound,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &testService{
				resp:     &GetResponse{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"},
				err:      nil,
				patchErr: tt.patchErr,
			}
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, tt.path, strings.NewReader(tt.body))
			r = r.WithContext(ctx.WithUserID(r.Context(), tt.userID))

			h.PatchSource(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
func TestHandlerGetAll(t *testing.T) {
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
			h := NewHandler(svc, slog.Default())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)

			h.GetAll(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
