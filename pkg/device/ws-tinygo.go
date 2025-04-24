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

func (l *wsLink) receive() (*Packet, error) {
	var data []byte
	var pkt Packet

	if err := websocket.Message.Receive(l.conn, &data); err != nil {
		return nil, err
	}

	l.Lock()
	l.lastRecv = time.Now()
	l.Unlock()

	if err := json.Unmarshal(data, &pkt); err != nil {
		return nil, fmt.Errorf("Unmarshalling error: %w", err)
	}
	return &pkt, nil
}

func (l *wsLink) receiveTimeout(timeout time.Duration) (*Packet, error) {
	l.conn.SetReadDeadline(time.Now().Add(timeout))
	pkt, err := l.receive()
	l.conn.SetReadDeadline(time.Time{})
	return pkt, err
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
	return l.send(&Packet{Path: "/ping"})
}

func (l *wsLink) receivePoll() (*Packet, error) {
	for {
		if l.timeToPing() {
			if err := l.sendPing(); err != nil {
				return nil, err
			}
		}
		pkt, err := l.receiveTimeout(time.Second)
		if err == nil {
			if pkt.Path == "/pong" {
				continue
			}
			return pkt, nil
		}
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			if l.timedOut() {
				return nil, err
			}
			continue
		}
		return nil, err
	}
	return nil, nil
}
