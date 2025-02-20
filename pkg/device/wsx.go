//go:build !tinygo

package device

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

// wsxHandle handles /wsx requests on an htmx WebSocket
func wsxHandle(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		LogError("Failed to upgrade to websocket", "error", err)
		return
	}
	defer ws.Close()

	wsxServe(ws, r)
}

// wsxServe handles htmx WebSocket connections
func wsxServe(ws *websocket.Conn, r *http.Request) {

	sessionId := r.URL.Query().Get("session-id")

	if sessionId == "hijack" {
		sessionId = sessionHijack()
	} else if sessionExpired(sessionId) {
		// Force full page refresh to start new session
		LogDebug("Session expired, refreshing", "id", sessionId)
		ws.WriteMessage(websocket.TextMessage, []byte("refresh"))
		return
	}

	sessionConn(sessionId, ws)

	// Force a refresh of root full view on successful ws connection, in
	// case anything has changed since last connect
	pkt := Packet{Dst: root.Id, Path: "/device"}
	sessionRoute(sessionId, &pkt)

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				LogError("Failed to read message", "error", err)
			}
			break
		}

		// The only message we're expecting is a ping from ws-send.
		// The ping message content:
		//
		// hdrs=map[HEADERS:map[
		//      HX-Current-URL:http://xxx:8000/
		//      HX-Request:true
		//      HX-Target:session
		//      HX-Trigger:session
		//      HX-Trigger-Name:<nil>
		//      session-id:98b87584-a6ad-4edb-8bab-faa32299a423]
		// ]

		var hdrs map[string]map[string]string
		if err := json.Unmarshal(message, &hdrs); err == nil {
			if hdr, ok := hdrs["HEADERS"]; ok {
				if sessionId, ok := hdr["session-id"]; ok {
					sessionKeepAlive(sessionId)
				}
			}
		}
	}

	sessionConn(sessionId, nil)
}
