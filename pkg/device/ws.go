//go:build !tinygo

package device

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pingPeriod  = 5 * time.Second
	pingTimeout = pingPeriod + time.Second
)

type wsLink struct {
	conn *websocket.Conn
	sync.RWMutex
	done bool
}

func (l *wsLink) Send(pkt *Packet) error {
	l.Lock()
	defer l.Unlock()
	return l.conn.WriteJSON(pkt)
}

func (l *wsLink) Close() {
	l.conn.Close()
}

func (s *server) receive(l *wsLink) (*Packet, error) {
	var pkt = s.newPacket()
	if err := l.conn.ReadJSON(&pkt); err != nil {
		l.done = true
		return nil, fmt.Errorf("Websocket read error: %w", err)
	}
	return pkt, nil
}

func (l *wsLink) setPongHandler() {
	l.conn.SetReadDeadline(time.Now().Add(pingTimeout))
	l.conn.SetPongHandler(func(appData string) error {
		l.conn.SetReadDeadline(time.Now().Add(pingTimeout))
		//println("Pong received, read deadline extended")
		return nil
	})
}

func (l *wsLink) startPing() {
	go func() {
		for {
			if l.done {
				return
			}
			l.Lock()
			if err := l.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				println("Ping error:", "err", err)
			}
			l.Unlock()
			time.Sleep(pingPeriod)
		}
	}()
}
