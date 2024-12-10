package device

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/gorilla/websocket"
)

type wsLink struct {
	conn     *websocket.Conn
	lastRecv time.Time
	lastSend time.Time
}

type announcement struct {
	Id           string
	Model        string
	Name         string
	DeployParams template.HTML
}

func (l *wsLink) Send(pkt *Packet) error {
	data, err := json.Marshal(pkt)
	if err != nil {
		return fmt.Errorf("Marshal error: %w", err)
	}
	if err := l.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("Send error: %w", err)
	}
	l.lastSend = time.Now()
	return nil
}

func (l *wsLink) Close() {
	l.conn.Close()
}

func (l *wsLink) receive() (*Packet, error) {
	_, data, err := l.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	l.lastRecv = time.Now()

	var pkt Packet
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

func (l *wsLink) receivePoll() (*Packet, error) {
	for {
		if time.Since(l.lastSend) >= pingDuration {
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
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return nil, err
		}
		if time.Since(l.lastRecv) > pingTimeout {
			return nil, err
		}
	}
}
