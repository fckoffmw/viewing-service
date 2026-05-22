package room

import (
	"errors"
	"log/slog"
	"testing"

	"w2g/internal/source"
)

var errRepo = errors.New("repo error")

type mockRoomRepo struct {
	rooms     []*Room
	getErr    error
	createErr error
	deleteErr error
	count     int
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

func (m *mockSourceGetter) GetSourceByID(id string) (*source.Source, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.src, nil
}

type mockHubManager struct {
	members int
}

func (m *mockHubManager) GetMembersOnline(roomID string) int {
	return m.members
}

func (m *mockHubManager) BroadcastSourceChanged(roomID, sourceID, sourceURL string) {
}

func (m *mockHubManager) Remove(roomID string) {
}

func TestServiceCreate(t *testing.T) {
	tests := []struct {
		name            string
		ownerID         string
		roomName        string
		wantErr         error
		maxRoomsPerUser int
		count           int
	}{
		{
			name:            "success",
			ownerID:         "user1",
			roomName:        "Movie Night",
			wantErr:         nil,
			maxRoomsPerUser: 10,
			count:           0,
		},
		{
			name:            "max rooms reached",
			ownerID:         "user1",
			roomName:        "Movie Night",
			wantErr:         ErrMaxRoomsReached,
			maxRoomsPerUser: 2,
			count:           2,
		},
		{
			name:            "unlimited rooms",
			ownerID:         "user1",
			roomName:        "Movie Night",
			wantErr:         nil,
			maxRoomsPerUser: 0,
			count:           5,
		},
		{
			name:            "no limit success",
			ownerID:         "user1",
			roomName:        "Movie Night",
			wantErr:         nil,
			maxRoomsPerUser: 0,
			count:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{count: tt.count}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, &mockHubManager{}, tt.maxRoomsPerUser)

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
		wantErr    error
	}{
		{
			name:       "found",
			inviteCode: "ABCD1234",
			wantErr:    nil,
		},
		{
			name:       "not found",
			inviteCode: "NOTEXIST",
			wantErr:    errors.New("room not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{
				rooms: []*Room{{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"}},
			}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, &mockHubManager{}, 10)

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
	srcGetter := &mockSourceGetter{src: &source.Source{ID: "s1", Name: "Test Video", URL: "http://example.com"}}
	svc := NewService(slog.Default(), repo, srcGetter, &mockHubManager{}, 10)

	setCurrentSourceIDForTest(svc, "1", "s1")

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

func setCurrentSourceIDForTest(svc *service, roomID, sourceID string) {
	svc.setCurrentSourceID(roomID, sourceID)
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name       string
		inviteCode string
		userID     string
		wantErr    error
		deleteErr  error
	}{
		{
			name:       "success as owner",
			inviteCode: "ABCD1234",
			userID:     "user1",
			wantErr:    nil,
			deleteErr:  nil,
		},
		{
			name:       "not owner",
			inviteCode: "ABCD1234",
			userID:     "user2",
			wantErr:    ErrNotOwner,
			deleteErr:  nil,
		},
		{
			name:       "not found",
			inviteCode: "NOTEXIST",
			userID:     "user1",
			wantErr:    errors.New("room not found"),
			deleteErr:  nil,
		},
		{
			name:       "delete error",
			inviteCode: "ABCD1234",
			wantErr:    errRepo,
			deleteErr:  errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{
				rooms:     []*Room{{ID: "1", InviteCode: "ABCD1234", OwnerID: "user1"}},
				deleteErr: tt.deleteErr,
			}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, &mockHubManager{}, 10)

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

func TestServiceGetAll(t *testing.T) {
	tests := []struct {
		name      string
		rooms     []*Room
		members   int
		wantCount int
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
			hub := &mockHubManager{members: tt.members}
			svc := NewService(slog.Default(), repo, &mockSourceGetter{}, hub, 10)

			result := svc.GetAll()

			if len(result) != tt.wantCount {
				t.Errorf("got %d rooms, want %d", len(result), tt.wantCount)
			}
			if tt.wantCount > 0 && result[0].MembersOnline != tt.members {
				t.Errorf("members = %d, want %d", result[0].MembersOnline, tt.members)
			}
		})
	}
}

func TestServicePatchSource(t *testing.T) {
	repo := &mockRoomRepo{
		rooms: []*Room{{ID: "1", Name: "Room1", InviteCode: "ABCD1234", OwnerID: "user1"}},
	}
	hub := &mockHubManager{}
	srcGetter := &mockSourceGetter{src: &source.Source{ID: "s1", Name: "Test", URL: "http://example.com"}}
	svc := NewService(slog.Default(), repo, srcGetter, hub, 10)

	err := svc.PatchSource("ABCD1234", "user1", "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svc.getCurrentSourceID("1") != "s1" {
		t.Errorf("expected current source s1, got %q", svc.getCurrentSourceID("1"))
	}
}

func TestServicePatchSourceNotOwner(t *testing.T) {
	repo := &mockRoomRepo{
		rooms: []*Room{{ID: "1", Name: "Room1", InviteCode: "ABCD1234", OwnerID: "user1"}},
	}
	svc := NewService(slog.Default(), repo, &mockSourceGetter{}, &mockHubManager{}, 10)

	err := svc.PatchSource("ABCD1234", "user2", "s1")
	if err != ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestServicePatchSourceNotFound(t *testing.T) {
	repo := &mockRoomRepo{
		rooms: []*Room{{ID: "1", Name: "Room1", InviteCode: "ABCD1234", OwnerID: "user1"}},
	}
	svc := NewService(slog.Default(), repo, &mockSourceGetter{err: errors.New("not found")}, &mockHubManager{}, 10)

	err := svc.PatchSource("ABCD1234", "user1", "s1")
	if err != ErrSourceNotFound {
		t.Errorf("expected ErrSourceNotFound, got %v", err)
	}
}

func TestServiceCurrentSourceID(t *testing.T) {
	svc := NewService(slog.Default(), &mockRoomRepo{}, &mockSourceGetter{}, &mockHubManager{}, 10)

	setCurrentSourceIDForTest(svc, "room1", "source1")
	setCurrentSourceIDForTest(svc, "room2", "source2")

	if svc.getCurrentSourceID("room1") != "source1" {
		t.Errorf("room1 source = %q, want %q", svc.getCurrentSourceID("room1"), "source1")
	}
	if svc.getCurrentSourceID("room2") != "source2" {
		t.Errorf("room2 source = %q, want %q", svc.getCurrentSourceID("room2"), "source2")
	}
	if svc.getCurrentSourceID("nonexistent") != "" {
		t.Errorf("nonexistent room source = %q, want empty", svc.getCurrentSourceID("nonexistent"))
	}
}
