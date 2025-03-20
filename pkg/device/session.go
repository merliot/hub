//go:build !tinygo

package device

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
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

var sessionsMax = 100

func newSessions() sessionMap {
	var sm sessionMap
	go sm.gcSessions()
	return sm
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

func (sm *sessionMap) noSessions(w http.ResponseWriter, r *http.Request) {
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
			//LogDebug("routeAll", "pkt", pkt)
			if err = s.renderPkt(pkt); err != nil {
				if err != errSessionNotConnected {
					return false
				}
			}
		}
		return true
	})
	return
}

func (sm *sessionMap) route(id string, pkt *Packet) error {
	if s, ok := sm.get(id); ok {
		//LogDebug("route", "pkt", pkt)
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
			LogError("sessionSend", "err", err)
		}
	}
}

/*
func (sm *sessionMap) sessionHijack() string {
	var hijackID string
	sm.drange(func(id string, s *session) bool {
		s.RLock()
		if s.conn != nil {
			hijackID = id
			s.RUnlock()
			return false // Stop iteration after finding one connected session
		}
		s.RUnlock()
		return true // Continue iteration
	})
	if hijackID != "" {
		return hijackID
	}
	return "none-to-hijack"
}
*/

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

func (sm *sessionMap) sortedAge() []string {

	keys := make([]string, 0)
	sm.drange(func(id string, s *session) bool {
		keys = append(keys, id)
		return true
	})

	// Sort keys based on lastUpdate, newest first
	sort.Slice(keys, func(i, j int) bool {
		s1, _ := sm.get(keys[i])
		s2, _ := sm.get(keys[j])
		return s1.lastUpdate.After(s2.lastUpdate)
	})

	return keys

}

type sessionStatus struct {
	Color  string
	Status string
}

func (sm *sessionMap) status() []sessionStatus {

	color := func(s *session) string {
		elapsed := time.Since(s.lastUpdate)
		switch {
		case elapsed < 30*time.Second:
			return "gold"
		case elapsed < 60*time.Second:
			return "orange"
		}
		return "red"
	}

	status := func(s *session) string {
		segs := strings.Split(s.id, "-")
		id := strings.ToUpper(segs[len(segs)-1])
		age := int(time.Since(s.lastUpdate).Truncate(time.Second).Seconds())
		connected := "C"
		if s.conn == nil {
			connected = " "
		}
		return fmt.Sprintf("%s %3d %s", id, age, connected)
	}

	var statuses = make([]sessionStatus, 0)
	for _, id := range sm.sortedAge() {
		s, _ := sm.get(id)
		s.RLock()
		statuses = append(statuses, sessionStatus{
			Color:  color(s),
			Status: status(s),
		})
		s.RUnlock()
	}

	return statuses
}

var errSessionNotConnected = errors.New("Session not connected")

func (s *session) _connected() bool {
	return s.conn != nil
}

func (s *session) connected() bool {
	s.RLock()
	defer s.RUnlock()
	return s._connected()
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
