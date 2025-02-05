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
	var pkt Packet
	if err := l.conn.ReadJSON(&pkt); err != nil {
		return nil, fmt.Errorf("ReadJSON error: %w", err)
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
	println("startPing")
	go func() {
		for {
			l.Lock()
			if err := l.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				//LogError("Ping error:", "err", err)
				l.Unlock()
				return
			}
			//LogInfo("Ping sent")
			l.Unlock()
			time.Sleep(pingDuration)
		}
	}()
}
