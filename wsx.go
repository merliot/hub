//go:build !tinygo

package hub

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/websocket"
)

// wsxHandle handles /wsx requests on an htmx WebSocket
func wsxHandle(w http.ResponseWriter, r *http.Request) {
	serv := websocket.Server{Handler: websocket.Handler(wsxServe)}
	serv.ServeHTTP(w, r)
}

// wsxServe handles htmx WebSocket connections
func wsxServe(ws *websocket.Conn) {

	defer ws.Close()

	req := ws.Request()
	sessionId := req.URL.Query().Get("session-id")
	if !sessionUpdate(sessionId) {
		LogError("Invalid session", "id", sessionId)
		return
	}

	sessionConn(sessionId, ws)

	// Force a refresh of root full view on successful ws connection, in
	// case anything has changed since last connect

	pkt := Packet{Dst: root.Id, Path: "/device"}
	sessionRoute(sessionId, &pkt)

	for {
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			//LogError("Can't receive", "err", err)
			break
		}

		// The only message we're expecting is a ping from ws-send.
		// The ping message content:
		//
		// hdrs=map[HEADERS:map[
		//	HX-Current-URL:http://xxx:8000/
		//	HX-Request:true
		//	HX-Target:session
		//	HX-Trigger:session
		//	HX-Trigger-Name:<nil>
		//	session-id:98b87584-a6ad-4edb-8bab-faa32299a423]
		// ]

		var hdrs map[string]map[string]string
		if err := json.Unmarshal([]byte(message), &hdrs); err == nil {
			if hdr, ok := hdrs["HEADERS"]; ok {
				if sessionId, ok := hdr["session-id"]; ok {
					sessionKeepAlive(sessionId)
				}
			}
		}
	}

	sessionConn(sessionId, nil)
}
