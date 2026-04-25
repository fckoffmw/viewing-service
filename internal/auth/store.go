package auth

import (
	"sync"
	"time"
)

type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *SessionStore) Set(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.SessionID] = session
}

func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[id]

	return sess, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func (s *SessionStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, sess := range s.sessions {
		if sess.ExpiresAt.Before(time.Now()) {
			delete(s.sessions, id)
		}
	}
}
