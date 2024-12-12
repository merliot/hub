//go:build !tinygo

// TODO get gorilla/websocket working on tinygo.  Currently hit:
//       ../../../go/pkg/mod/github.com/gorilla/websocket@v1.5.1/client.go:18:2: package net/http/httptrace is not in std (/root/...

package device

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsLink struct {
	conn *websocket.Conn
	sync.Mutex
}

type announcement struct {
	Id           string
	Model        string
	Name         string
	DeployParams template.HTML
}

func (l *wsLink) Send(pkt *Packet) error {
	l.Lock()
	defer l.Unlock()
	return l.conn.WriteJSON(pkt)
}

func (l *wsLink) Close() {
	l.conn.Close()
}

var pingDuration = 4 * time.Second

// var pingTimeout = pingDuration + time.Second
// TODO allow two ping periods before timing out, rather than one, to
// workaround some issue I'm having deploying to cloud where ping (or pong)
// packet's are getting buffered and timing out.
var pingTimeout = 2*pingDuration + time.Second

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

func (l *wsLink) receive() (*Packet, error) {
	var pkt Packet
	if err := l.conn.ReadJSON(&pkt); err != nil {
		return nil, fmt.Errorf("ReadJSON error: %w", err)
	}
	return &pkt, nil
}
