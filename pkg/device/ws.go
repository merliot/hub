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
	lastSend time.Time
}

func (l *wsLink) Send(pkt *Packet) error {
	data, err := json.Marshal(pkt)
	if err != nil {
		return fmt.Errorf("Marshal error: %w", err)
	}
	l.Lock()
	defer l.Unlock()
	if err := websocket.Message.Send(l.conn, data); err != nil {
		return fmt.Errorf("Send error: %w", err)
	}
	l.lastSend = time.Now()
	return nil
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
		LogError("Unmarshal Error", "data", string(data))
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
var pingTimeout = 2*pingDuration + time.Second

func (l *wsLink) timeToPing() bool {
	l.RLock()
	defer l.RUnlock()
	return time.Since(l.lastSend) >= pingDuration
}

func (l *wsLink) timedOut() bool {
	l.RLock()
	defer l.RUnlock()
	return time.Since(l.lastRecv) > pingTimeout
}

func (l *wsLink) receivePoll() (*Packet, error) {
	for {
		if l.timeToPing() {
			if err := l.Send(&Packet{Path: "/ping"}); err != nil {
				return nil, err
			}
		}
		pkt, err := l.receiveTimeout(time.Second)
		if err == nil {
			if pkt.Path == "/ping" {
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
