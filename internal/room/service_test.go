package room

import (
	"errors"
	"testing"

	"w2g/internal/source"
)

var errRepo = errors.New("repo error")

type mockRepo struct {
	room   *Room
	source *source.Source
	err    error
}

func (m *mockRepo) GetGlobalRoom() (*Room, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.room, nil
}

func (m *mockRepo) GetSourceById(id string) (*source.Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.source, nil
}

func (m *mockRepo) UpdateGlobalRoomSource(srcID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return srcID, nil
}

func TestServiceGetGlobalRoom(t *testing.T) {
	tests := []struct {
		name    string
		room    *Room
		wantID  string
		wantErr error
	}{
		{
			name:    "success",
			room:    &Room{ID: "1", SourceID: "10"},
			wantID:  "1",
			wantErr: nil,
		},
		{
			name:    "repo error",
			room:    nil,
			wantID:  "",
			wantErr: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{room: tt.room, err: tt.wantErr}
			svc := NewService(repo)

			got, err := svc.GetGlobalRoom()

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.wantID {
				t.Errorf("got.ID = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}

func TestServiceGetSourceById(t *testing.T) {
	src := &source.Source{ID: "1", Name: "Film", Url: "http://vk.com/1"}

	tests := []struct {
		name     string
		source   *source.Source
		wantName string
		wantErr  error
	}{
		{
			name:     "success",
			source:   src,
			wantName: "Film",
			wantErr:  nil,
		},
		{
			name:     "repo error",
			source:   nil,
			wantName: "",
			wantErr:  errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{source: tt.source, err: tt.wantErr}
			svc := NewService(repo)

			got, err := svc.GetSourceById("1")

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.wantName {
				t.Errorf("got.Name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

func TestServiceUpdateGlobalRoomSource(t *testing.T) {
	tests := []struct {
		name    string
		srcID   string
		wantID  string
		wantErr error
	}{
		{
			name:    "success",
			srcID:   "5",
			wantID:  "5",
			wantErr: nil,
		},
		{
			name:    "repo error",
			srcID:   "",
			wantID:  "",
			wantErr: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{err: tt.wantErr}
			svc := NewService(repo)

			got, err := svc.UpdateGlobalRoomSource(tt.srcID)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantID {
				t.Errorf("got = %q, want %q", got, tt.wantID)
			}
		})
	}
}
