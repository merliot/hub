//go:build !tinygo

package device

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsLink struct {
	name string
	conn *websocket.Conn
	sync.RWMutex
	done bool
}

var (
	pingPeriod  = 2 * time.Second
	pingTimeout = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins, update as necessary
		return true
	},
}

func (l *wsLink) Name() string {
	return l.name
}

func (l *wsLink) Send(pkt *Packet) error {
	l.Lock()
	defer l.Unlock()
	return l.conn.WriteJSON(pkt)
}

func (l *wsLink) RemoteAddr() net.Addr {
	return l.conn.RemoteAddr()
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
