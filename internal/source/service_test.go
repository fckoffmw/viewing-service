package source

import (
	"errors"
	"testing"
)

var errRepo = errors.New("repo error")

type mockRepo struct {
	sources []Source
	err     error
}

func (m *mockRepo) GetAllSources() ([]Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}

func TestServiceGetAllSources(t *testing.T) {
	tests := []struct {
		name      string
		sources   []Source
		wantLen   int
		wantFirst string
		wantErr   error
	}{
		{
			name:      "empty",
			sources:   nil,
			wantLen:   0,
			wantFirst: "",
			wantErr:   nil,
		},
		{
			name: "one source",
			sources: []Source{
				{ID: "1", Name: "Film", Url: "http://vk.com/1"},
			},
			wantLen:   1,
			wantFirst: "Film",
			wantErr:   nil,
		},
		{
			name: "multiple sources",
			sources: []Source{
				{ID: "1", Name: "Film1", Url: "http://vk.com/1"},
				{ID: "2", Name: "Film2", Url: "http://vk.com/2"},
			},
			wantLen:   2,
			wantFirst: "Film1",
			wantErr:   nil,
		},
		{
			name:      "repo error",
			sources:   nil,
			wantLen:   0,
			wantFirst: "",
			wantErr:   errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{sources: tt.sources, err: tt.wantErr}
			svc := NewService(repo)

			got, err := svc.GetAllSources()

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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
