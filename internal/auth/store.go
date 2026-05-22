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
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if !sess.ExpiresAt.Before(time.Now()) {
		return sess, true
	}

	s.mu.Lock()
	sess, ok = s.sessions[id]
	if ok && sess.ExpiresAt.Before(time.Now()) {
		delete(s.sessions, id)
		s.mu.Unlock()

		return nil, false
	}
	s.mu.Unlock()

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
