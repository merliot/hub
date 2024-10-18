//go:build !tinygo

package hub

import (
	"fmt"
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
		fmt.Println("Invalid session", sessionId)
		return
	}

	sessionConn(sessionId, ws)

	// Force a refresh of root full view on successful ws connection, in
	// case anything has changed since last connect

	pkt := Packet{Dst: root.Id, Path: "/device"}
	sessionRoute(sessionId, &pkt)

	// We use htmx websockets in one-direction only, from the server to the
	// client, and only used for sending HTML snippets back to the client.
	//
	// Keep the websocket connection open, waiting for receives (which will
	// never come, see above).  Break on EOF or other error.

	for {
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			fmt.Println("Can't receive:", err)
			break
		}
		fmt.Println("Received message from client:", message)
	}

	sessionConn(sessionId, nil)
}
