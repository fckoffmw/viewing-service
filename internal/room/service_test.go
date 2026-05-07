package room

import (
	"errors"
	"log/slog"
	"testing"

	"w2g/internal/source"
)

// Helper to get private currentSourceID map for testing
func getCurrentSourceIDForTest(roomID string) string {
	return currentSourceID[roomID]
}

var errRepo = errors.New("repo error")

type mockRoomRepo struct {
	rooms       []*Room
	getErr     error
	createErr  error
	deleteErr  error
	count      int
}

func (m *mockRoomRepo) Create(room *Room) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.rooms = append(m.rooms, room)
	return nil
}

func (m *mockRoomRepo) GetByInviteCode(inviteCode string) (*Room, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, r := range m.rooms {
		if r.InviteCode == inviteCode {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockRoomRepo) GetByID(id string) (*Room, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, r := range m.rooms {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockRoomRepo) Delete(inviteCode string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	var newRooms []*Room
	for _, r := range m.rooms {
		if r.InviteCode != inviteCode {
			newRooms = append(newRooms, r)
		}
	}
	m.rooms = newRooms
	return nil
}

func (m *mockRoomRepo) CountByOwnerID(ownerID string) int {
	return m.count
}

func (m *mockRoomRepo) GetAll() []*Room {
	return m.rooms
}

type mockSourceGetter struct {
	src *source.Source
	err error
}

func (m *mockSourceGetter) GetSourceById(id string) (*source.Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.src, nil
}

type mockHubGetter struct {
	members int
}

func (m *mockHubGetter) GetMembersOnline(roomID string) int {
	return m.members
}

func TestServiceCreate(t *testing.T) {
	tests := []struct {
		name          string
		ownerID       string
		roomName      string
		wantErr       error
		maxRoomsPerUser int
		count        int
	}{
		{
			name:          "success",
			ownerID:      "user1",
			roomName:     "Movie Night",
			wantErr:      nil,
			maxRoomsPerUser: 10,
			count:         0,
		},
		{
			name:          "max rooms reached",
			ownerID:      "user1",
			roomName:     "Movie Night",
			wantErr:      ErrMaxRoomsReached,
			maxRoomsPerUser: 2,
			count:        2,
		},
		{
			name:          "unlimited rooms",
			ownerID:      "user1",
			roomName:     "Movie Night",
			wantErr:      nil,
			maxRoomsPerUser: 0,
			count:        5,
		},
		{
			name:          "no limit success",
			ownerID:      "user1",
			roomName:     "Movie Night",
			wantErr:      nil,
			maxRoomsPerUser: 0,
			count:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{count: tt.count}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, nil, tt.maxRoomsPerUser)

			resp, err := svc.Create(CreateRequest{Name: tt.roomName}, tt.ownerID)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Name != tt.roomName {
				t.Errorf("resp.Name = %q, want %q", resp.Name, tt.roomName)
			}
			if resp.OwnerID != tt.ownerID {
				t.Errorf("resp.OwnerID = %q, want %q", resp.OwnerID, tt.ownerID)
			}
		})
	}
}

func TestServiceGetByInviteCode(t *testing.T) {
	tests := []struct {
		name       string
		inviteCode string
		wantErr   error
	}{
		{
			name:       "found",
			inviteCode: "ABCD1234",
			wantErr:   nil,
		},
		{
			name:       "not found",
			inviteCode: "NOTEXIST",
			wantErr:   errors.New("room not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{
				rooms: []*Room{{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"}},
			}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, nil, 10)

			resp, err := svc.GetByInviteCode(tt.inviteCode)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.InviteCode != tt.inviteCode {
				t.Errorf("resp.InviteCode = %q, want %q", resp.InviteCode, tt.inviteCode)
			}
		})
	}
}

func TestServiceGetByInviteCodeWithSource(t *testing.T) {
	repo := &mockRoomRepo{
		rooms: []*Room{{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"}},
	}
	srcGetter := &mockSourceGetter{src: &source.Source{ID: "s1", Name: "Test Video", Url: "http://example.com"}}
	svc := NewService(slog.Default(), repo, srcGetter, nil, 10)

	svc.SetCurrentSourceID("1", "s1")

	resp, err := svc.GetByInviteCode("ABCD1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CurrentSource == nil {
		t.Error("expected current source, got nil")
	}
	if resp.CurrentSource.Name != "Test Video" {
		t.Errorf("current source name = %q, want %q", resp.CurrentSource.Name, "Test Video")
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name       string
		inviteCode string
		userID     string
		wantErr   error
		deleteErr error
	}{
		{
			name:       "success as owner",
			inviteCode: "ABCD1234",
			userID:    "user1",
			wantErr:   nil,
			deleteErr: nil,
		},
		{
			name:       "not owner",
			inviteCode: "ABCD1234",
			userID:    "user2",
			wantErr:   ErrNotOwner,
			deleteErr: nil,
		},
		{
			name:       "not found",
			inviteCode: "NOTEXIST",
			userID:    "user1",
			wantErr:   errors.New("room not found"),
			deleteErr: nil,
		},
		{
			name:       "delete error",
			inviteCode: "ABCD1234",
			wantErr:   errRepo,
			deleteErr: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{
				rooms: []*Room{{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"}},
				deleteErr: tt.deleteErr,
			}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, nil, 10)

			err := svc.Delete(tt.inviteCode, tt.userID)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceGetAllRooms(t *testing.T) {
	tests := []struct {
		name         string
		rooms        []*Room
		members      int
		wantCount    int
	}{
		{
			name:      "empty",
			rooms:     []*Room{},
			members:   0,
			wantCount: 0,
		},
		{
			name:      "single room",
			rooms:     []*Room{{ID: "1", Name: "Room1", InviteCode: "R1", OwnerID: "user1"}},
			members:   2,
			wantCount: 1,
		},
		{
			name:      "multiple rooms",
			rooms:     []*Room{{ID: "1", Name: "Room1", InviteCode: "R1", OwnerID: "user1"}, {ID: "2", Name: "Room2", InviteCode: "R2", OwnerID: "user2"}},
			members:   5,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{rooms: tt.rooms}
			hub := &mockHubGetter{members: tt.members}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, hub, 10)

			result := svc.GetAllRooms()

			if len(result) != tt.wantCount {
				t.Errorf("got %d rooms, want %d", len(result), tt.wantCount)
			}
			if tt.wantCount > 0 && result[0].MembersOnline != tt.members {
				t.Errorf("members = %d, want %d", result[0].MembersOnline, tt.members)
			}
		})
	}
}

func TestServiceSetCurrentSourceID(t *testing.T) {
	repo := &mockRoomRepo{}
	svc := NewService(slog.Default(), repo, &mockSourceGetter{}, nil, 10)

	svc.SetCurrentSourceID("room1", "source1")
	svc.SetCurrentSourceID("room2", "source2")

	if svc.GetCurrentSourceID("room1") != "source1" {
		t.Errorf("room1 source = %q, want %q", svc.GetCurrentSourceID("room1"), "source1")
	}
	if svc.GetCurrentSourceID("room2") != "source2" {
		t.Errorf("room2 source = %q, want %q", svc.GetCurrentSourceID("room2"), "source2")
	}
	if svc.GetCurrentSourceID("nonexistent") != "" {
		t.Errorf("nonexistent room source = %q, want empty", svc.GetCurrentSourceID("nonexistent"))
	}
}