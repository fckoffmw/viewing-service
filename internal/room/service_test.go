package room

import (
	"errors"
	"testing"
)

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
			svc := NewService(repo, tt.maxRoomsPerUser)

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
			svc := NewService(repo, 10)

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
			svc := NewService(repo, 10)

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