//go:build !tinygo

package device

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type lastView struct {
	view  string
	level int
}

type session struct {
	sessionId  string
	conn       *websocket.Conn
	lastUpdate time.Time
	lastViews  map[string]lastView // key: device Id
	rwMutex
}

var sessions = make(map[string]*session) // key: sessionId
var sessionsMu rwMutex
var sessionCount int32
var sessionCountMax = int32(1000)

func init() {
	go gcSessions()
}

func _newSession(sessionId string, conn *websocket.Conn) *session {
	return &session{
		sessionId:  sessionId,
		conn:       conn,
		lastUpdate: time.Now(),
		lastViews:  make(map[string]lastView),
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
	sessionCount++

	return sessionId, true
}

func sessionConn(sessionId string, conn *websocket.Conn) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		s.Lock()
		s.conn = conn
		s.lastUpdate = time.Now()
		s.Unlock()
	}
}

func sessionExpired(sessionId string) bool {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	_, ok := sessions[sessionId]
	return !ok
}

func sessionUpdate(sessionId string) bool {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		s.Lock()
		defer s.Unlock()
		if s.connected() {
			s.lastUpdate = time.Now()
			return true
		}
		return false
	}

	// Session expired
	return false
}

func sessionKeepAlive(sessionId string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		s.Lock()
		s.lastUpdate = time.Now()
		s.Unlock()
	}
}

func sessionSave(sessionId, deviceId, view string, level int) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		s.Lock()
		s.lastUpdate = time.Now()
		lastView := s.lastViews[deviceId]
		lastView.view = view
		lastView.level = level
		s.lastViews[deviceId] = lastView
		s.Unlock()
	}
}

func sessionLastView(sessionId, deviceId string) (view string, level int, err error) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	s, ok := sessions[sessionId]
	if !ok {
		return "", 0, fmt.Errorf("Invalid session %s", sessionId)
	}

	s.RLock()
	defer s.RUnlock()

	v, ok := s.lastViews[deviceId]
	if !ok {
		return "", 0, fmt.Errorf("Session %s: invalid device Id %s", sessionId, deviceId)
	}
	return v.view, v.level, nil
}

var errSessionNotConnected = errors.New("Session not connected")

func (s *session) connected() bool {
	return s.conn != nil
}

func (s *session) send(data []byte) error {
	s.Lock()
	defer s.Unlock()
	if s.connected() {
		return s.conn.WriteMessage(websocket.TextMessage, data)
	}
	return errSessionNotConnected
}

func (s *session) renderPkt(pkt *Packet) error {
	s.RLock()
	if !s.connected() {
		s.RUnlock()
		return errSessionNotConnected
	}
	s.RUnlock()
	var buf bytes.Buffer
	if err := deviceRenderPkt(&buf, s.sessionId, pkt); err != nil {
		return err
	}
	return s.send(buf.Bytes())
}

func sessionsRoute(pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, s := range sessions {
		if pkt.SessionId == "" || pkt.SessionId == s.sessionId {
			LogDebug("sessionsRoute", "pkt", pkt)
			if err := s.renderPkt(pkt); err != nil {
				if err != errSessionNotConnected {
					LogError("sessionsRoute", "err", err)
				}
			}
		}
	}
}

func sessionRoute(sessionId string, pkt *Packet) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		// LogDebug("sessionRoute", "pkt", pkt)
		if err := s.renderPkt(pkt); err != nil {
			if err != errSessionNotConnected {
				LogError("sessionRoute", "err", err)
			}
		}
	}
}

func sessionSend(sessionId, htmlSnippet string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if s, ok := sessions[sessionId]; ok {
		if err := s.send([]byte(htmlSnippet)); err != nil {
			LogError("sessionSend", "err", err)
		}
	}
}

func sessionHijack() string {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for _, s := range sessions {
		s.RLock()
		if s.conn != nil {
			s.RUnlock()
			return s.sessionId
		}
		s.RUnlock()
	}

	return "none-to-hijack"
}

func gcSessions() {
	minute := 1 * time.Minute
	ticker := time.NewTicker(minute)
	defer ticker.Stop()
	for range ticker.C {
		sessionsMu.Lock()
		for sessionId, s := range sessions {
			s.RLock()
			if time.Since(s.lastUpdate) > minute {
				delete(sessions, sessionId)
				sessionCount--
			}
			s.RUnlock()
		}
		sessionsMu.Unlock()
	}
}

func sessionsSortedAge() []string {
	keys := make([]string, 0, len(sessions))
	for key := range sessions {
		keys = append(keys, key)
	}

	// Sort keys based on lastUpdate, newest first
	sort.Slice(keys, func(i, j int) bool {
		return sessions[keys[i]].lastUpdate.After(sessions[keys[j]].lastUpdate)
	})

	return keys
}

func (s session) sortedViewIds() []string {
	s.RLock()
	defer s.RUnlock()

	keys := make([]string, 0, len(s.lastViews))
	for id := range s.lastViews {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	return keys
}

type sessionStatus struct {
	Color  string
	Status string
}

func sessionsStatus() []sessionStatus {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

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
		segs := strings.Split(s.sessionId, "-")
		id := strings.ToUpper(segs[len(segs)-1])
		age := int(time.Since(s.lastUpdate).Truncate(time.Second).Seconds())
		connected := "C"
		if s.conn == nil {
			connected = " "
		}
		views := ""
		for _, id := range s.sortedViewIds() {
			view := s.lastViews[id]
			views = views + " " + strings.ToUpper(string(view.view[0])) + strconv.Itoa(view.level)
		}
		return fmt.Sprintf("%s %3d %s %s", id, age, connected, views)
	}

	var statuses = make([]sessionStatus, 0, len(sessions))
	for _, id := range sessionsSortedAge() {
		s := sessions[id]
		statuses = append(statuses, sessionStatus{
			Color:  color(s),
			Status: status(s),
		})
	}

	return statuses
}
