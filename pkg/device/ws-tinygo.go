//go:build tinygo

package device

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type wsLink struct {
	name string
	conn *websocket.Conn
	sync.RWMutex
	lastRecv time.Time
	lastPing time.Time
}

func (l *wsLink) send(pkt *Packet) error {
	data, err := json.Marshal(pkt)
	if err != nil {
		return fmt.Errorf("Marshal error: %w", err)
	}
	if err := websocket.Message.Send(l.conn, data); err != nil {
		return fmt.Errorf("Send error: %w", err)
	}
	return nil
}

func (l *wsLink) Name() string {
	return l.name
}

func (l *wsLink) Send(pkt *Packet) error {
	l.Lock()
	defer l.Unlock()
	return l.send(pkt)
}

func (l *wsLink) RemoteAddr() net.Addr {
	return l.conn.RemoteAddr()
}

func (l *wsLink) Close() {
	l.conn.Close()
}

func (l *wsLink) receive(pkt *Packet) error {
	var data []byte

	if err := websocket.Message.Receive(l.conn, &data); err != nil {
		return err
	}

	l.Lock()
	l.lastRecv = time.Now()
	l.Unlock()

	if err := json.Unmarshal(data, pkt); err != nil {
		return fmt.Errorf("Unmarshalling error: %w", err)
	}
	return nil
}

func (l *wsLink) receiveTimeout(pkt *Packet, timeout time.Duration) error {
	l.conn.SetReadDeadline(time.Now().Add(timeout))
	err := l.receive(pkt)
	l.conn.SetReadDeadline(time.Time{})
	return err
}

var pingDuration = 4 * time.Second
var pingTimeout = 4*pingDuration + time.Second

func (l *wsLink) timeToPing() bool {
	l.RLock()
	defer l.RUnlock()
	return time.Since(l.lastPing) >= pingDuration
}

func (l *wsLink) timedOut() bool {
	l.RLock()
	defer l.RUnlock()
	return time.Since(l.lastRecv) > pingTimeout
}

func (l *wsLink) sendPing() error {
	l.Lock()
	defer l.Unlock()
	l.lastPing = time.Now()
	return l.send(&Packet{Path: "ping"})
}

func (l *wsLink) readJSON(pkt *Packet) error {
	for {
		if l.timeToPing() {
			if err := l.sendPing(); err != nil {
				return err
			}
		}
		err := l.receiveTimeout(pkt, time.Second)
		if err == nil {
			if pkt.Path == "pong" {
				continue
			}
			return nil
		}
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			if l.timedOut() {
				return err
			}
			continue
		}
		return err
	}
}

func (s *server) receive(l *wsLink) (*Packet, error) {
	var pkt = s.newPacket()
	if err := l.readJSON(pkt); err != nil {
		return nil, fmt.Errorf("Websocket read error: %w", err)
	}
	return pkt, nil
}

func (s *server) receiveTimeout(l *wsLink, timeout time.Duration) (*Packet, error) {
	var pkt = s.newPacket()
	err := l.receiveTimeout(pkt, timeout)
	return pkt, err
}
