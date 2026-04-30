package auth

import (
	"sync"
	"testing"
	"time"
)

func TestNewSessionStore(t *testing.T) {
	store := NewSessionStore(time.Minute)
	if store == nil {
		t.Fatal("expected store, got nil")
	}
	if store.sessions == nil {
		t.Fatal("expected sessions map, got nil")
	}
}

func TestSessionStore_Set_Get(t *testing.T) {
	store := NewSessionStore(time.Minute)

	sess := &Session{
		SessionID:  "test-session-1",
		UserID:    "user-1",
		CreatedAt: time.Now(),
		LastSeenAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	store.Set(sess)

	got, ok := store.Get("test-session-1")
	if !ok {
		t.Fatal("expected session, got not found")
	}
	if got.UserID != "user-1" {
		t.Errorf("expected UserID=user-1, got %s", got.UserID)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	store := NewSessionStore(time.Minute)

	_, ok := store.Get("nonexistent")
	if ok {
		t.Fatal("expected not found, got session")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore(time.Minute)

	sess := &Session{
		SessionID:  "test-session-1",
		UserID:    "user-1",
		CreatedAt: time.Now(),
		LastSeenAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	store.Set(sess)

	store.Delete("test-session-1")

	_, ok := store.Get("test-session-1")
	if ok {
		t.Fatal("expected session deleted, got session")
	}
}

func TestSessionStore_Cleanup(t *testing.T) {
	store := NewSessionStore(time.Minute)

	expired := &Session{
		SessionID:  "expired-session",
		UserID:    "user-1",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		LastSeenAt: time.Now().Add(-24 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	valid := &Session{
		SessionID:  "valid-session",
		UserID:    "user-2",
		CreatedAt: time.Now(),
		LastSeenAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	store.Set(expired)
	store.Set(valid)

	store.cleanup()

	_, ok := store.Get("expired-session")
	if ok {
		t.Error("expected expired session to be deleted")
	}

	_, ok = store.Get("valid-session")
	if !ok {
		t.Error("expected valid session to remain")
	}
}

func TestSessionStore_Concurrency(t *testing.T) {
	store := NewSessionStore(time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			store.Set(&Session{
				SessionID:  "session-" + string(rune(id)),
				UserID:    "user",
				CreatedAt: time.Now(),
				LastSeenAt: time.Now(),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			})
		}(i)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			store.Get("session-" + string(rune(id)))
		}(i)
	}

	wg.Wait()
}