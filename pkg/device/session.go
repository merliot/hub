//go:build !tinygo

package device

import (
	"bytes"
	_ "embed"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type session struct {
	id         string
	conn       *websocket.Conn
	lastUpdate time.Time
	rwMutex
}

type sessionMap struct {
	sync.Map // key: session id, value: *session
}

var errSessionNotConnected = errors.New("Session not connected")

var sessionsMax = 100

func (sm *sessionMap) start() {
	go sm.gcSessions()
}

func (sm *sessionMap) get(id string) (*session, bool) {
	value, ok := sm.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*session), true
}

func (sm *sessionMap) drange(f func(string, *session) bool) {
	sm.Range(func(key, value any) bool {
		id := key.(string)
		s := value.(*session)
		return f(id, s)
	})
}

func (sm *sessionMap) length() int {
	count := 0
	sm.drange(func(string, *session) bool {
		count++
		return true
	})
	return count
}

func (sm *sessionMap) newSession() (string, bool) {

	if sm.length() >= sessionsMax {
		return "", false
	}

	sessionId := uuid.New().String()
	s := &session{id: sessionId, lastUpdate: time.Now()}
	sm.Store(sessionId, s)

	return sessionId, true
}

func (sm *sessionMap) noSessions(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "no more sessions", http.StatusTooManyRequests)
}

func (sm *sessionMap) setConn(id string, conn *websocket.Conn) {
	if s, ok := sm.get(id); ok {
		s.Lock()
		s.conn = conn
		s.lastUpdate = time.Now()
		s.Unlock()
	}
}

func (sm *sessionMap) expired(id string) bool {
	_, exists := sm.get(id)
	return !exists
}

func (sm *sessionMap) keepAlive(id string) {
	if s, ok := sm.get(id); ok {
		s.Lock()
		s.lastUpdate = time.Now()
		s.Unlock()
	}
}

func (sm *sessionMap) routeAll(pkt *Packet) (err error) {
	sm.drange(func(id string, s *session) bool {
		if pkt.SessionId == "" || pkt.SessionId == id {
			if err = s.renderPkt(pkt); err != nil {
				if err != errSessionNotConnected {
					return false
				}
				err = nil
			}
		}
		return true
	})
	return
}

func (sm *sessionMap) route(id string, pkt *Packet) error {
	if s, ok := sm.get(id); ok {
		if err := s.renderPkt(pkt); err != nil {
			if err != errSessionNotConnected {
				return err
			}
		}
	}
	return nil
}

func (sm *sessionMap) send(id, htmlSnippet string) {
	if s, ok := sm.get(id); ok {
		if err := s.send([]byte(htmlSnippet)); err != nil {
		}
	}
}

func (sm *sessionMap) gcSessions() {
	minute := 1 * time.Minute
	ticker := time.NewTicker(minute)
	defer ticker.Stop()
	for range ticker.C {
		sm.drange(func(id string, s *session) bool {
			s.mu.RLock()
			if time.Since(s.lastUpdate) > minute {
				sm.Delete(id)
			}
			s.RUnlock()
			return true
		})
	}
}

type sessionStatus struct {
	Color  string
	Status string
}

func (s *session) _connected() bool {
	return s.conn != nil
}

// _send wants s.Lock to serialize writes on socket
func (s *session) _send(data []byte) error {
	if s._connected() {
		return s.conn.WriteMessage(websocket.TextMessage, data)
	}
	return errSessionNotConnected
}

func (s *session) send(data []byte) error {
	s.Lock()
	defer s.Unlock()
	return s._send(data)
}

func (s *session) renderPkt(pkt *Packet) error {

	// Using Lock rather than RLock because _send needs Lock
	s.Lock()
	defer s.Unlock()

	if !s._connected() {
		return errSessionNotConnected
	}

	var buf bytes.Buffer
	if err := pkt.render(&buf, s.id); err != nil {
		return err
	}

	return s._send(buf.Bytes())
}
