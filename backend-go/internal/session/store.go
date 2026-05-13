package session

import (
	"context"
	"log"
	"sync"
	"time"
)

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	audit    []AuditEntry
}

func NewStore() *Store {
	return &Store{sessions: make(map[string]*Session)}
}

func (s *Store) Put(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ID] = sess
}

// Get returns a session, lazily marking it expired if its TTL passed while pending.
// Returns (nil, false) if not found.
func (s *Store) Get(id string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	if sess.Status == StatusPending && nowMillis() > sess.ExpiresAt {
		sess.Status = StatusExpired
		s.logLocked(AuditEntry{Ts: nowMillis(), SessionID: sess.ID, Event: "session.expired"})
	}
	return sess, true
}

// Update applies a function under the write lock and returns the updated session.
func (s *Store) Update(id string, mutate func(*Session)) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	mutate(sess)
	return sess, true
}

func (s *Store) Log(entry AuditEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logLocked(entry)
}

func (s *Store) logLocked(entry AuditEntry) {
	s.audit = append(s.audit, entry)
	log.Printf("[audit] %s %s sid=%s %v",
		time.UnixMilli(entry.Ts).UTC().Format(time.RFC3339),
		entry.Event, entry.SessionID, entry.Detail)
}

func (s *Store) Audit() []AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AuditEntry, len(s.audit))
	copy(out, s.audit)
	return out
}

// RunSweeper expires pending sessions past their TTL every 5s until ctx is done.
func (s *Store) RunSweeper(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sweepOnce()
		}
	}
}

func (s *Store) sweepOnce() {
	now := nowMillis()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sess := range s.sessions {
		if sess.Status == StatusPending && now > sess.ExpiresAt {
			sess.Status = StatusExpired
			s.logLocked(AuditEntry{Ts: now, SessionID: sess.ID, Event: "session.expired"})
		}
	}
}

func nowMillis() int64 { return time.Now().UnixMilli() }
