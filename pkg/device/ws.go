//go:build !tinygo

package device

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var pingDuration = 2 * time.Second

// var pingTimeout = pingDuration + time.Second
// TODO allow two ping periods before timing out, rather than one, to
// workaround some issue I'm having deploying to cloud where ping (or pong)
// packet's are getting buffered and timing out.
var pingTimeout = 4*pingDuration + time.Second

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

func (l *wsLink) receive() (*Packet, error) {
	if err := l.conn.ReadJSON(&pkt); err != nil {
		l.done = true
		return nil, fmt.Errorf("Websocket read error: %w", err)
	}
	return &pkt, nil
}

func (l *wsLink) setPongHandler() {
	l.conn.SetReadDeadline(time.Now().Add(pingTimeout))
	l.conn.SetPongHandler(func(appData string) error {
		l.conn.SetReadDeadline(time.Now().Add(pingTimeout))
		//LogInfo("Pong received, read deadline extended")
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
				LogError("Ping error:", "err", err)
			}
			l.Unlock()
			time.Sleep(pingDuration)
		}
	}()
}
