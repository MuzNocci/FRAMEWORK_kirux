package session

import (
	"encoding/hex"
	"kyrux/core/security/crypton"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	ID      string
	Values  map[string]any
	Expires time.Time
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

func NewStore(ttl time.Duration) *Store {
	s := &Store{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	go s.gc()
	return s
}

func (s *Store) New() (*Session, error) {
	b, err := crypton.RandomBytes(32)
	if err != nil {
		return nil, err
	}
	id := hex.EncodeToString(b)
	sess := &Session{
		ID:      id,
		Values:  make(map[string]any),
		Expires: time.Now().Add(s.ttl),
	}
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
	return sess, nil
}

func (s *Store) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok || time.Now().After(sess.Expires) {
		return nil, false
	}
	return sess, true
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func (s *Store) gc() {
	ticker := time.NewTicker(s.ttl)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for id, sess := range s.sessions {
			if time.Now().After(sess.Expires) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

func CookieName() string { return "kyrux_session" }

// SetCookie define o cookie de sessão com as flags de segurança corretas.
// secure deve ser true em produção (HTTPS). Usar em conjunto com session.New().
func SetCookie(w http.ResponseWriter, sessionID string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName(),
		Value:    sessionID,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

func FromRequest(r *http.Request, store *Store) (*Session, bool) {
	cookie, err := r.Cookie(CookieName())
	if err != nil {
		return nil, false
	}
	return store.Get(cookie.Value)
}
