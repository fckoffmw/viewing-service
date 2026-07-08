package repo

import (
	"os"
	"testing"
	"time"

	"w2g/internal/auth"
	"w2g/internal/room"
)

type testRow struct {
	ID   string
	Name string
}

type testRow4 struct {
	ID   string
	Name string
	Url  string
	Tag  string
}

type rowChecker interface {
	check(t *testing.T, index int, expected rowChecker)
}

func (r testRow) check(t *testing.T, index int, expected rowChecker) {
	exp := expected.(testRow)
	if r.ID != exp.ID {
		t.Errorf("row %d: ID expected %q, got %q", index, exp.ID, r.ID)
	}
	if r.Name != exp.Name {
		t.Errorf("row %d: Name expected %q, got %q", index, exp.Name, r.Name)
	}
}

func (r testRow4) check(t *testing.T, index int, expected rowChecker) {
	exp := expected.(testRow4)
	if r.ID != exp.ID {
		t.Errorf("row %d: ID expected %q, got %q", index, exp.ID, r.ID)
	}
	if r.Name != exp.Name {
		t.Errorf("row %d: Name expected %q, got %q", index, exp.Name, r.Name)
	}
	if r.Url != exp.Url {
		t.Errorf("row %d: Url expected %q, got %q", index, exp.Url, r.Url)
	}
	if r.Tag != exp.Tag {
		t.Errorf("row %d: Tag expected %q, got %q", index, exp.Tag, r.Tag)
	}
}

type testCase struct {
	name     string
	rows     [][]string
	expected []rowChecker
}

func TestRowsTo(t *testing.T) {
	tests := []testCase{
		{
			name:     "header only",
			rows:     [][]string{{"id", "name"}},
			expected: nil,
		},
		{
			name: "one row",
			rows: [][]string{
				{"id", "name"},
				{"1", "test"},
			},
			expected: []rowChecker{testRow{ID: "1", Name: "test"}},
		},
		{
			name: "multiple rows",
			rows: [][]string{
				{"id", "name"},
				{"1", "first"},
				{"2", "second"},
			},
			expected: []rowChecker{
				testRow{ID: "1", Name: "first"},
				testRow{ID: "2", Name: "second"},
			},
		},
		{
			name: "four fields",
			rows: [][]string{
				{"id", "name", "url", "tag"},
				{"1", "film", "http://a", "movie"},
				{"2", "show", "http://b", "series"},
			},
			expected: []rowChecker{
				testRow4{ID: "1", Name: "film", Url: "http://a", Tag: "movie"},
				testRow4{ID: "2", Name: "show", Url: "http://b", Tag: "series"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expected) == 0 {
				got, err := rowsTo[testRow](tt.rows)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(got) != 0 {
					t.Fatalf("expected 0 rows, got %d", len(got))
				}
				return
			}

			switch tt.expected[0].(type) {
			case testRow:
				got, err := rowsTo[testRow](tt.rows)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				for i := range got {
					got[i].check(t, i, tt.expected[i])
				}
			case testRow4:
				got, err := rowsTo[testRow4](tt.rows)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				for i := range got {
					got[i].check(t, i, tt.expected[i])
				}
			}
		})
	}
}

func TestRowsTo_Errors(t *testing.T) {
	type intRow struct {
		ID int
	}

	_, err := rowsTo[intRow]([][]string{
		{"id"},
		{"not_a_number"},
	})
	if err == nil {
		t.Fatal("expected error for invalid int, got nil")
	}
}

func setupTestStorage(t *testing.T, files map[string]string) *csvStorage {
	t.Helper()

	dir := t.TempDir() + "/"

	// сначала создаём хранилище (пустые csv)
	storage, err := NewCSVStorage(dir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// потом перезаписываем тестовыми данными
	for name, content := range files {
		if err := os.WriteFile(dir+name, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", name, err)
		}
	}

	return storage
}

func TestNewCSVStorage(t *testing.T) {
	dir := t.TempDir() + "/"

	_, err := NewCSVStorage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// проверяем что файлы созданы
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 csv files, got %d", len(files))
	}
}

func TestNewCSVStorage_InvalidPath(t *testing.T) {
	_, err := NewCSVStorage("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestGetAllSources(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"sources.csv": "id,name,url\n1,Seven,http://vk.com/seven\n2,Matrix,http://vk.com/matrix\n",
		"rooms.csv":   "id,source_id\n1,1\n",
	})

	sources, err := storage.GetAllSources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}

	if sources[0].ID != "1" || sources[0].Name != "Seven" {
		t.Errorf("expected first source ID=1, Name=Seven, got ID=%s, Name=%s", sources[0].ID, sources[0].Name)
	}
}

func TestGetSourceByID(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"sources.csv": "id,name,url\n1,Seven,http://vk.com/seven\n2,Matrix,http://vk.com/matrix\n",
		"rooms.csv":   "id,source_id\n1,1\n",
	})

	t.Run("found", func(t *testing.T) {
		s, err := storage.GetSourceByID("2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name != "Matrix" {
			t.Errorf("expected Matrix, got %s", s.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := storage.GetSourceByID("999")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetAllRooms(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"sources.csv": "id,name,url\n",
		"rooms.csv":   "id,name,owner_id,invite_code,created_at\n1,Test,user1,ABCD1234,2025-01-01T00:00:00Z\n",
	})

	rooms, err := storage.GetAllRooms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rooms) != 1 {
		t.Errorf("expected 1 room, got %d", len(rooms))
	}
	if rooms[0].ID != "1" {
		t.Errorf("expected room ID=1, got %s", rooms[0].ID)
	}
	if rooms[0].Name != "Test" {
		t.Errorf("expected name=Test, got %s", rooms[0].Name)
	}
}

func TestAddRoom(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"sources.csv": "id,name,url\n",
		"rooms.csv":   "id,name,owner_id,invite_code,created_at\n1,First,user1,ABCD1234,2025-01-01T00:00:00Z\n",
	})

	room := room.Room{
		Name:       "Second",
		OwnerID:    "user2",
		InviteCode:  "EFGH5678",
		CreatedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	id, err := storage.AddRoom(&room)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "1" {
		t.Errorf("expected id=1, got %s", id)
	}

	rooms, err := storage.GetAllRooms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
}

func TestUpdateRoom(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"sources.csv": "id,name,url\n",
		"rooms.csv":   "id,name,owner_id,invite_code,created_at\n1,OldName,user1,ABCD1234,2025-01-01T00:00:00Z\n",
	})

	room := room.Room{
		ID:         "1",
		Name:       "NewName",
		OwnerID:    "user1",
		InviteCode: "ABCD1234",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := storage.UpdateRoom(room)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rooms, err := storage.GetAllRooms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rooms[0].Name != "NewName" {
		t.Errorf("expected name=NewName, got %s", rooms[0].Name)
	}
}

func TestDeleteRoom(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"sources.csv": "id,name,url\n",
		"rooms.csv":   "id,name,owner_id,invite_code,created_at\n1,First,user1,ABCD1234,2025-01-01T00:00:00Z\n2,Second,user2,EFGH5678,2025-01-02T00:00:00Z\n",
	})

	err := storage.DeleteRoom("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rooms, err := storage.GetAllRooms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 1 {
		t.Errorf("expected 1 room, got %d", len(rooms))
	}
	if rooms[0].ID != "2" {
		t.Errorf("expected remaining room ID=2, got %s", rooms[0].ID)
	}
}

func TestGetUserByUsername(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"users.csv":   "id,username,password_hash,created_at\n1,alice,$2a$12$hash," + time.Now().Format(time.RFC3339) + "\n2,bob,$2a$12$hash2," + time.Now().Add(time.Hour).Format(time.RFC3339) + "\n",
		"sources.csv": "id,name,url\n",
		"rooms.csv":   "id,source_id\n1,1\n",
	})

	t.Run("found", func(t *testing.T) {
		user, err := storage.GetUserByUsername("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user == nil {
			t.Fatal("expected user, got nil")
		}
		if user.Username != "alice" {
			t.Errorf("expected username=alice, got %s", user.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		user, err := storage.GetUserByUsername("unknown")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil, got %v", user)
		}
	})
}

func TestAddUser(t *testing.T) {
	storage := setupTestStorage(t, map[string]string{
		"users.csv":   "id,username,password_hash,created_at\n1,alice,$2a$12$hash," + time.Now().Format(time.RFC3339) + "\n",
		"sources.csv": "id,name,url\n",
		"rooms.csv":   "id,source_id\n1,1\n",
	})

	user := auth.User{
		Username:     "newuser",
		PasswordHash: "$2a$12$newtesthash",
	}

	id, err := storage.AddUser(&user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "1" {
		t.Errorf("expected id=1, got %s", id)
	}

	// проверяем что пользователь добавлен
	newUser, err := storage.GetUserByUsername("newuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newUser == nil {
		t.Fatal("expected user, got nil")
	}
	if newUser.Username != "newuser" {
		t.Errorf("expected username=newuser, got %s", newUser.Username)
	}
}
