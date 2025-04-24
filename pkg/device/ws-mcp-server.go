//go:build !tinygo

package device

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// wsMcpHandle handles /wsmcp requests from MCP servers
func (s *server) wsMcpHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logError("Upgrading WebSocket", "err", err)
		return
	}
	s.wsMcpServe(conn)
}

func (s *server) wsMcpServe(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{name: "/wsmcp server", conn: conn}

	s.logDebug("Adding MCP Uplink")
	s.uplinks.add(link)

	// Start ping/pong
	link.setPongHandler()
	link.startPing()

	// Just keep the connection alive - MCP servers only receive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	s.logDebug("Removing MCP Uplink")
	s.uplinks.remove(link)
	link.done = true
}
