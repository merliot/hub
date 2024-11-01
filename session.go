//go:build !tinygo

package hub

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ietxaniz/delock"
	"golang.org/x/net/websocket"
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
	delock.RWMutex
}

var sessions = make(map[string]*session) // key: sessionId
var sessionsMu delock.RWMutex
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
	lockId, err := sessionsMu.Lock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.Unlock(lockId)

	if sessionCount >= sessionCountMax {
		return "", false
	}

	sessionId := uuid.New().String()
	sessions[sessionId] = _newSession(sessionId, nil)
	sessionCount += 1

	return sessionId, true
}

func sessionConn(sessionId string, conn *websocket.Conn) {

	lockId, err := sessionsMu.Lock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.Unlock(lockId)

	if s, ok := sessions[sessionId]; ok {
		s.conn = conn
		s.lastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, conn)
		sessionCount += 1
	}
}

func sessionUpdate(sessionId string) bool {

	lockId, err := sessionsMu.Lock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.Unlock(lockId)

	if s, ok := sessions[sessionId]; ok {
		s.lastUpdate = time.Now()
		return true
	}

	// Session expired
	return false
}

func sessionKeepAlive(sessionId string) bool {

	lockId, err := sessionsMu.Lock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.Unlock(lockId)

	if s, ok := sessions[sessionId]; ok {
		s.lastUpdate = time.Now()
		data, _ := json.Marshal(&Packet{Path: "/ping"})
		if s.conn != nil {
			websocket.Message.Send(s.conn, string(data))
		}
		return true
	}

	// Session expired
	return false
}

func _sessionSave(sessionId, deviceId, view string, level int) {

	if s, ok := sessions[sessionId]; ok {
		lockId, err := s.Lock()
		if err != nil {
			panic(err)
		}
		defer s.Unlock(lockId)
		s.lastUpdate = time.Now()
		lastView := s.lastViews[deviceId]
		lastView.view = view
		lastView.level = level
		s.lastViews[deviceId] = lastView
	}
}

func _sessionLastView(sessionId, deviceId string) (string, int, error) {
	s, ok := sessions[sessionId]
	if !ok {
		return "", 0, fmt.Errorf("Invalid session %s", sessionId)
	}

	lockId, err := s.RLock()
	if err != nil {
		panic(err)
	}
	defer s.RUnlock(lockId)

	view, ok := s.lastViews[deviceId]
	if !ok {
		return "", 0, fmt.Errorf("Session %s: invalid device Id %s", sessionId, deviceId)
	}
	return view.view, view.level, nil
}

func sessionLastView(sessionId, deviceId string) (view string, level int, err error) {
	lockId, err := sessionsMu.RLock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.RUnlock(lockId)
	return _sessionLastView(sessionId, deviceId)
}

func (s session) _renderPkt(pkt *Packet) {
	var buf bytes.Buffer
	if err := deviceRenderPkt(&buf, s.sessionId, pkt); err != nil {
		LogError("Rendering pkt", "err", err)
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionsRoute(pkt *Packet) {

	lockId, err := sessionsMu.RLock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.RUnlock(lockId)

	for _, s := range sessions {
		if s.conn != nil {
			if pkt.SessionId == "" || pkt.SessionId == s.sessionId {
				//LogInfo("SessionsRoute", "pkt", pkt)
				s._renderPkt(pkt)
			}
		}
	}
}

func sessionRoute(sessionId string, pkt *Packet) {

	lockId, err := sessionsMu.RLock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.RUnlock(lockId)

	if s, ok := sessions[sessionId]; ok {
		if s.conn != nil {
			//LogInfo("SessionRoute", "pkt", pkt)
			s._renderPkt(pkt)
		}
	}
}

func sessionSend(sessionId, htmlSnippet string) {

	lockId, err := sessionsMu.RLock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.RUnlock(lockId)

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
		lockId, err := sessionsMu.Lock()
		if err != nil {
			panic(err)
		}
		for sessionId, s := range sessions {
			if time.Since(s.lastUpdate) > minute {
				delete(sessions, sessionId)
				sessionCount -= 1
			}
		}
		sessionsMu.Unlock(lockId)
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
	lockId, err := s.RLock()
	if err != nil {
		panic(err)
	}
	defer s.RUnlock(lockId)

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
	lockId, err := sessionsMu.RLock()
	if err != nil {
		panic(err)
	}
	defer sessionsMu.RUnlock(lockId)

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

	var statuses = make([]sessionStatus, len(sessions))
	for _, id := range sessionsSortedAge() {
		s := sessions[id]
		statuses = append(statuses, sessionStatus{
			Color:  color(s),
			Status: status(s),
		})
	}

	return statuses
}
