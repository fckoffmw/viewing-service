package source

import (
	"errors"
	"strconv"
	"testing"
)

var errRepo = errors.New("repo error")

type mockRepo struct {
	sources []Source
	nextID  int
	err     error
}

func (m *mockRepo) GetAllSources() ([]Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}

func (m *mockRepo) AddSource(s Source) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.nextID++
	id := strconv.Itoa(m.nextID)
	s.ID = id
	m.sources = append(m.sources, s)
	return id, nil
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

func TestServiceAddSource(t *testing.T) {
	tests := []struct {
		name    string
		sources []Source
		newName string
		newURL  string
		wantID  string
		wantErr error
	}{
		{
			name:    "success",
			sources: []Source{{ID: "1", Name: "Film", Url: "http://vk.com/1"}},
			newName: "New Film",
			newURL:  "http://vk.com/2",
			wantID:  "2",
			wantErr: nil,
		},
		{
			name:    "empty sources",
			sources: nil,
			newName: "New Film",
			newURL:  "http://vk.com/2",
			wantID:  "1",
			wantErr: nil,
		},
		{
			name:    "repo error",
			sources: nil,
			newName: "New Film",
			newURL:  "http://vk.com/2",
			wantID:  "",
			wantErr: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{sources: tt.sources, nextID: len(tt.sources), err: tt.wantErr}
			svc := NewService(repo)

			got, err := svc.AddSource(tt.newName, tt.newURL)

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
			if len(repo.sources) != len(tt.sources)+1 {
				t.Errorf("sources len = %d, want %d", len(repo.sources), len(tt.sources)+1)
			}
		})
	}
}
