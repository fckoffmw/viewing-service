package room

import (
	"errors"
	"testing"
)

var errCSV = errors.New("csv error")

type mockCSVStorage struct {
	rooms  []Room
	getErr error
	addErr error
	delErr error
}

func (m *mockCSVStorage) GetAllRooms() ([]Room, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.rooms, nil
}

func (m *mockCSVStorage) AddRoom(r *Room) (string, error) {
	if m.addErr != nil {
		return "", m.addErr
	}
	r.ID = "1"
	return r.ID, nil
}

func (m *mockCSVStorage) UpdateRoom(r Room) error {
	return nil
}

func (m *mockCSVStorage) DeleteRoom(id string) error {
	if m.delErr != nil {
		return m.delErr
	}
	var newRooms []Room
	for _, rm := range m.rooms {
		if rm.ID != id {
			newRooms = append(newRooms, rm)
		}
	}
	m.rooms = newRooms
	return nil
}

func TestNewStore(t *testing.T) {
	tests := []struct {
		name  string
		rooms []Room
		want  int
		err   error
	}{
		{
			name:  "success",
			rooms: []Room{{ID: "1", Name: "Room1", InviteCode: "R1"}, {ID: "2", Name: "Room2", InviteCode: "R2"}},
			want:  2,
			err:   nil,
		},
		{
			name:  "empty",
			rooms: []Room{},
			want:  0,
			err:   nil,
		},
		{
			name:  "csv error",
			rooms: nil,
			want:  0,
			err:   errCSV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csv := &mockCSVStorage{rooms: tt.rooms, getErr: tt.err}
			store, err := NewStore(csv)

			if tt.err != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(store.rooms) != tt.want {
				t.Errorf("len(store.rooms) = %d, want %d", len(store.rooms), tt.want)
			}
		})
	}
}

func TestStoreGetByInviteCode(t *testing.T) {
	store := &store{
		rooms: map[string]*Room{
			"ABCD1234": {ID: "1", Name: "Test", InviteCode: "ABCD1234"},
		},
		storage: nil,
	}

	tests := []struct {
		name       string
		inviteCode string
		wantID     string
		wantErr    error
	}{
		{
			name:       "found",
			inviteCode: "ABCD1234",
			wantID:     "1",
			wantErr:    nil,
		},
		{
			name:       "not found",
			inviteCode: "NOTEXIST",
			wantID:     "",
			wantErr:    errors.New("room not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := store.GetByInviteCode(tt.inviteCode)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if room.ID != tt.wantID {
				t.Errorf("room.ID = %q, want %q", room.ID, tt.wantID)
			}
		})
	}
}

func TestStoreGetByID(t *testing.T) {
	store := &store{
		rooms: map[string]*Room{
			"ABCD1234": {ID: "1", Name: "Test", InviteCode: "ABCD1234"},
		},
		storage: nil,
	}

	tests := []struct {
		name    string
		id     string
		wantID string
		wantErr error
	}{
		{
			name:    "found",
			id:     "1",
			wantID: "1",
			wantErr: nil,
		},
		{
			name:    "not found",
			id:     "999",
			wantID: "",
			wantErr: errors.New("room not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := store.GetByID(tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if room.ID != tt.wantID {
				t.Errorf("room.ID = %q, want %q", room.ID, tt.wantID)
			}
		})
	}
}

func TestStoreGetByOwnerID(t *testing.T) {
	store := &store{
		rooms: map[string]*Room{
			"A": {ID: "1", OwnerID: "user1", InviteCode: "A"},
			"B": {ID: "2", OwnerID: "user1", InviteCode: "B"},
			"C": {ID: "3", OwnerID: "user2", InviteCode: "C"},
		},
		storage: nil,
	}

	tests := []struct {
		name      string
		ownerID   string
		wantCount int
	}{
		{
			name:      "multiple rooms",
			ownerID:   "user1",
			wantCount: 2,
		},
		{
			name:      "single room",
			ownerID:   "user2",
			wantCount: 1,
		},
		{
			name:      "no rooms",
			ownerID:   "user3",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms, err := store.GetByOwnerID(tt.ownerID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(rooms) != tt.wantCount {
				t.Errorf("len(rooms) = %d, want %d", len(rooms), tt.wantCount)
			}
		})
	}
}

func TestStoreDelete(t *testing.T) {
	csv := &mockCSVStorage{}
	store := &store{
		rooms: map[string]*Room{
			"ABCD1234": {ID: "1", InviteCode: "ABCD1234"},
		},
		storage: csv,
	}

	tests := []struct {
		name       string
		inviteCode string
		wantErr   error
		deleteErr error
	}{
		{
			name:       "success",
			inviteCode: "ABCD1234",
			wantErr:    nil,
			deleteErr: nil,
		},
		{
			name:       "not found",
			inviteCode: "NOTEXIST",
			wantErr:   errors.New("room not found"),
			deleteErr: nil,
		},
		{
			name:       "csv error",
			inviteCode: "ABCD1234",
			wantErr:   errors.New("csv error"),
			deleteErr: errors.New("csv error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.rooms = map[string]*Room{"ABCD1234": {ID: "1", InviteCode: "ABCD1234"}}
			csv.delErr = tt.deleteErr

			err := store.Delete(tt.inviteCode)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, ok := store.rooms[tt.inviteCode]; ok {
				t.Error("room still exists after delete")
			}
		})
	}
}

func TestStoreCountByOwnerID(t *testing.T) {
	store := &store{
		rooms: map[string]*Room{
			"A": {ID: "1", OwnerID: "user1", InviteCode: "A"},
			"B": {ID: "2", OwnerID: "user1", InviteCode: "B"},
			"C": {ID: "3", OwnerID: "user2", InviteCode: "C"},
		},
		storage: nil,
	}

	tests := []struct {
		name      string
		ownerID   string
		wantCount int
	}{
		{
			name:      "two rooms",
			ownerID:   "user1",
			wantCount: 2,
		},
		{
			name:      "one room",
			ownerID:   "user2",
			wantCount: 1,
		},
		{
			name:      "no rooms",
			ownerID:   "user3",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := store.CountByOwnerID(tt.ownerID)
			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}