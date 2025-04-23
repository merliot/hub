//go:build !tinygo

package device

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

// mcpWsDial connects to a Hub's /wsmcp endpoint to receive notifications
func (ms *MCPServer) mcpWsDial() error {

	var hdr = http.Header{}

	if ms.url == "" {
		return fmt.Errorf("Missing dial URL")
	}

	// Transform http(s) URL to ws(s) URL
	wsURL, err := url.Parse(ms.url)
	if err != nil {
		return fmt.Errorf("Invalid URL: %w", err)
	}
	switch wsURL.Scheme {
	case "http":
		wsURL.Scheme = "ws"
	case "https":
		wsURL.Scheme = "wss"
	default:
		return fmt.Errorf("Unsupported URL scheme: %s", wsURL.Scheme)
	}

	// Add /wsmcp to the path
	wsURL.Path = "/wsmcp"

	// If valid user, set the basic auth header for the request
	if ms.user != "" {
		req, err := http.NewRequest("GET", wsURL.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.SetBasicAuth(ms.user, ms.passwd)
		hdr = req.Header
	}

	// Connect to the Hub with custom headers
	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), hdr)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	// Service the MCP websocket client
	go ms.mcpWsClient(conn)

	return nil
}

func (ms *MCPServer) receive(l *wsLink) (*Packet, error) {
	var pkt = &Packet{}
	if err := l.conn.ReadJSON(&pkt); err != nil {
		l.done = true
		return nil, fmt.Errorf("Websocket read error: %w", err)
	}
	return pkt, nil
}

func (ms *MCPServer) mcpWsClient(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{conn: conn}

	link.setPongHandler()
	link.startPing()

	// Receive packets from Hub and convert to MCP notifications
	for {
		pkt, err := ms.receive(link)
		if err != nil {
			break
		}

		var msgContent map[string]any
		if err := json.Unmarshal(pkt.Msg, &msgContent); err != nil {
			continue
		}

		params := map[string]any{
			"dst_id": pkt.Dst,
			"path":   pkt.Path,
			"msg":    msgContent,
		}

		ms.MCPServer.SendNotificationToAllClients("notification/update", params)
	}

	link.done = true
}
