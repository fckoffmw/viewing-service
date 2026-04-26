package auth

import (
	"sync"
	"time"
)

type sessionStore struct {
	sessions        map[string]*Session
	cleanupInterval time.Duration
	mu              sync.RWMutex
	stopCh          chan struct{}
}

func NewSessionStore(cleanupInterval time.Duration) *sessionStore {
	return &sessionStore{
		sessions:        make(map[string]*Session),
		cleanupInterval: cleanupInterval,
		stopCh:          make(chan struct{}),
	}
}

func (s *sessionStore) CleanupLoop() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCh:
			return
		}
	}
}

func (s *sessionStore) Set(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.SessionID] = session
}

func (s *sessionStore) Get(id string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}

	if sess.ExpiresAt.Before(time.Now()) {
		delete(s.sessions, id)
		return nil, false
	}

	return sess, true
}

func (s *sessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func (s *sessionStore) Stop() {
	close(s.stopCh)
}

func (s *sessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for id, sess := range s.sessions {
		if sess.ExpiresAt.Before(now) {
			delete(s.sessions, id)
		}
	}
}
