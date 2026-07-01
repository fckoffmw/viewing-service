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

func (m *mockRepo) GetSourceByID(id string) (*Source, error) {
	if m.err != nil {
		return nil, m.err
	}

	for i := range m.sources {
		if m.sources[i].ID == id {
			return &m.sources[i], nil
		}
	}

	return nil, nil
}

func (m *mockRepo) UpdateSource(s *Source) error {
	if m.err != nil {
		return m.err
	}

	for i := range m.sources {
		if m.sources[i].ID == s.ID {
			m.sources[i].Name = s.Name
			m.sources[i].URL = s.URL

			return nil
		}
	}

	return errRepo
}

func (m *mockRepo) AddSource(s *Source) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.nextID++
	id := strconv.Itoa(m.nextID)
	s.ID = id
	m.sources = append(m.sources, *s)
	return id, nil
}

func (m *mockRepo) DeleteSource(id string) error {
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

func TestServiceGetAll(t *testing.T) {
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
				{ID: "1", Name: "Film", URL: "http://vk.com/1"},
			},
			wantLen:   1,
			wantFirst: "Film",
			wantErr:   nil,
		},
		{
			name: "multiple sources",
			sources: []Source{
				{ID: "1", Name: "Film1", URL: "http://vk.com/1"},
				{ID: "2", Name: "Film2", URL: "http://vk.com/2"},
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

			got, err := svc.GetAll()

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

func TestServiceAdd(t *testing.T) {
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
			sources: []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
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

			got, err := svc.Add(tt.newName, tt.newURL)

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

func TestServiceUpdate(t *testing.T) {
	tests := []struct {
		name     string
		sources  []Source
		updateID string
		newName  string
		newURL   string
		mockErr  error
		wantErr  bool
		wantName string
	}{
		{
			name:     "success",
			sources:  []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			updateID: "1",
			newName:  "Updated Film",
			newURL:   "http://vk.com/new",
			wantErr:  false,
			wantName: "Updated Film",
		},
		{
			name:     "not found",
			sources:  []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			updateID: "999",
			newName:  "Updated",
			newURL:   "http://vk.com/new",
			wantErr:  true,
		},
		{
			name:     "repo error",
			sources:  []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			updateID: "1",
			newName:  "Updated",
			newURL:   "http://vk.com/new",
			mockErr:  errRepo,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{sources: tt.sources, err: tt.mockErr}
			svc := NewService(repo)

			err := svc.Update(tt.updateID, tt.newName, tt.newURL)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(repo.sources) > 0 && repo.sources[0].Name != tt.wantName {
				t.Errorf("source.Name = %q, want %q", repo.sources[0].Name, tt.wantName)
			}
		})
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name      string
		sources   []Source
		deleteID  string
		wantErr   error
		wantLen   int
	}{
		{
			name:     "success",
			sources:  []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			deleteID: "1",
			wantErr:  nil,
			wantLen:  0,
		},
		{
			name:     "repo error",
			sources:  []Source{{ID: "1", Name: "Film", URL: "http://vk.com/1"}},
			deleteID: "999",
			wantErr:  errRepo,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{sources: tt.sources, err: tt.wantErr}
			svc := NewService(repo)

			err := svc.Delete(tt.deleteID)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(repo.sources) != tt.wantLen {
				t.Errorf("sources len = %d, want %d", len(repo.sources), tt.wantLen)
			}
		})
	}
}
