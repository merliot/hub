//go:build !tinygo

package device

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
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

var sessions sync.Map // key: sessionId
var sessionCount int32
var sessionCountMax = int32(1000)

func init() {
	go gcSessions()
}

func newSession() (string, bool) {

	if sessionCount >= sessionCountMax {
		return "", false
	}

	sessionId := uuid.New().String()
	s := &session{id: sessionId, lastUpdate: time.Now()}
	sessions.Store(sessionId, s)
	sessionCount++

	return sessionId, true
}

func sessionConn(sessionId string, conn *websocket.Conn) {
	if v, ok := sessions.Load(sessionId); ok {
		s := v.(*session)
		s.Lock()
		s.conn = conn
		s.lastUpdate = time.Now()
		s.Unlock()
	}
}

func sessionExpired(sessionId string) bool {
	_, exists := sessions.Load(sessionId)
	return !exists
}

func sessionKeepAlive(sessionId string) {
	if v, ok := sessions.Load(sessionId); ok {
		s := v.(*session)
		s.Lock()
		s.lastUpdate = time.Now()
		s.Unlock()
	}
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
	if err := deviceRenderPkt(&buf, s.id, pkt); err != nil {
		return err
	}

	return s._send(buf.Bytes())
}

func sessionsRoute(pkt *Packet) {
	sessions.Range(func(_, value any) bool {
		s := value.(*session)
		if pkt.SessionId == "" || pkt.SessionId == s.id {
			//LogDebug("sessionsRoute", "pkt", pkt)
			if err := s.renderPkt(pkt); err != nil {
				if err != errSessionNotConnected {
					LogError("sessionsRoute", "err", err)
				}
			}
		}
		return true
	})
}

func sessionRoute(sessionId string, pkt *Packet) {
	if v, ok := sessions.Load(sessionId); ok {
		s := v.(*session)
		//LogDebug("sessionRoute", "pkt", pkt)
		if err := s.renderPkt(pkt); err != nil {
			if err != errSessionNotConnected {
				LogError("sessionRoute", "err", err)
			}
		}
	}
}

func sessionSend(sessionId, htmlSnippet string) {
	if v, ok := sessions.Load(sessionId); ok {
		s := v.(*session)
		if err := s.send([]byte(htmlSnippet)); err != nil {
			LogError("sessionSend", "err", err)
		}
	}
}

func sessionHijack() string {
	var hijackID string
	sessions.Range(func(_, value interface{}) bool {
		s := value.(*session)
		s.RLock()
		if s.conn != nil {
			hijackID = s.id
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

func gcSessions() {
	minute := 1 * time.Minute
	ticker := time.NewTicker(minute)
	defer ticker.Stop()
	for range ticker.C {
		sessions.Range(func(key, value interface{}) bool {
			sessionId := key.(string)
			s := value.(*session)
			s.mu.RLock()
			if time.Since(s.lastUpdate) > minute {
				sessions.Delete(sessionId)
				gcViews(sessionId)
				sessionCount--
			}
			s.RUnlock()
			return true
		})
	}
}

func sessionsSortedAge() []string {
	keys := make([]string, 0)
	sessions.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})

	// Sort keys based on lastUpdate, newest first
	sort.Slice(keys, func(i, j int) bool {
		s1, _ := sessions.Load(keys[i])
		s2, _ := sessions.Load(keys[j])
		return s1.(*session).lastUpdate.After(s2.(*session).lastUpdate)
	})

	return keys

}

type sessionStatus struct {
	Color  string
	Status string
}

func sessionsStatus() []sessionStatus {

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
	for _, id := range sessionsSortedAge() {
		v, _ := sessions.Load(id)
		s := v.(*session)
		s.RLock()
		statuses = append(statuses, sessionStatus{
			Color:  color(s),
			Status: status(s),
		})
		s.RUnlock()
	}

	return statuses
}
