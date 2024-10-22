//go:build !tinygo

package hub

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

//go:embed template/sessions.tmpl
var sessionsTemplate string

type lastView struct {
	View  string
	Level int
}

type session struct {
	sessionId    string
	conn         *websocket.Conn
	LastUpdate   time.Time
	LastViews    map[string]lastView // key: device Id
	sync.RWMutex `json:"-"`
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex
var sessionCount int32
var sessionCountMax = int32(1000)

func init() {
	go gcSessions()
}

func _newSession(sessionId string, conn *websocket.Conn) *session {
	return &session{
		sessionId:  sessionId,
		conn:       conn,
		LastUpdate: time.Now(),
		LastViews:  make(map[string]lastView),
	}
}

func newSession() (string, bool) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if sessionCount >= sessionCountMax {
		return "", false
	}

	sessionId := uuid.New().String()
	sessions[sessionId] = _newSession(sessionId, nil)
	sessionCount += 1

	return sessionId, true
}

func (s session) Age() string {
	age := time.Since(s.LastUpdate).Truncate(time.Second).Seconds()
	return fmt.Sprintf("%03d", int(age))
}

func sessionConn(sessionId string, conn *websocket.Conn) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if s, ok := sessions[sessionId]; ok {
		s.conn = conn
		s.LastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, conn)
		sessionCount += 1
	}
}

func sessionUpdate(sessionId string) bool {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if s, ok := sessions[sessionId]; ok {
		s.LastUpdate = time.Now()
		return true
	}

	// Session expired
	return false
}

func sessionKeepAlive(sessionId string) bool {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if s, ok := sessions[sessionId]; ok {
		s.LastUpdate = time.Now()
		data, _ := json.Marshal(&Packet{Path: "/ping"})
		websocket.Message.Send(s.conn, string(data))
		return true
	}

	// Session expired
	return false
}

func _sessionSave(sessionId, deviceId, view string, level int) {

	if s, ok := sessions[sessionId]; ok {
		s.Lock()
		defer s.Unlock()
		s.LastUpdate = time.Now()
		lastView := s.LastViews[deviceId]
		lastView.View = view
		lastView.Level = level
		s.LastViews[deviceId] = lastView
	}
}

func _sessionLastView(sessionId, deviceId string) (string, int, error) {
	s, ok := sessions[sessionId]
	if !ok {
		return "", 0, fmt.Errorf("Invalid session %s", sessionId)
	}

	s.RLock()
	defer s.RUnlock()

	view, ok := s.LastViews[deviceId]
	if !ok {
		return "", 0, fmt.Errorf("Session %s: invalid device Id %s", sessionId, deviceId)
	}
	return view.View, view.Level, nil
}

func sessionLastView(sessionId, deviceId string) (view string, level int, err error) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	return _sessionLastView(sessionId, deviceId)
}

func (s session) _renderPkt(pkt *Packet) {
	var buf bytes.Buffer
	if err := deviceRenderPkt(&buf, s.sessionId, pkt); err != nil {
		fmt.Println("\nError rendering pkt:", err, "\n")
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionsRoute(pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, s := range sessions {
		if s.conn != nil {
			//fmt.Println("=== sessionsRoute", pkt)
			s._renderPkt(pkt)
		}
	}
}

func sessionRoute(sessionId string, pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		if s.conn != nil {
			//fmt.Println("=== sessionRoute", pkt)
			s._renderPkt(pkt)
		}
	}
}

func sessionSend(sessionId, htmlSnippet string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		if s.conn != nil {
			websocket.Message.Send(s.conn, htmlSnippet)
		}
	}
}

func gcSessions() {
	minute := 1 * time.Minute
	ticker := time.NewTicker(minute)
	defer ticker.Stop()
	for range ticker.C {
		sessionsMu.Lock()
		for sessionId, s := range sessions {
			if time.Since(s.LastUpdate) > minute {
				delete(sessions, sessionId)
				sessionCount -= 1
			}
		}
		sessionsMu.Unlock()
	}
}
