//go:build tinygo

package device

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"
)

func newConfig(wsUrl *url.URL, user, passwd string) (*websocket.Config, error) {

	// Set the origin to match the WebSocket server’s scheme and host
	origin := &url.URL{Scheme: "http", Host: wsUrl.Host}
	if wsUrl.Scheme == "wss" {
		origin.Scheme = "https"
	}

	// Configure the websocket
	config, err := websocket.NewConfig(wsUrl.String(), origin.String())
	if err != nil {
		return nil, err
	}

	// If valid user, set the basic auth header for the request
	if user != "" {
		req, err := http.NewRequest("GET", wsUrl.String(), nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, passwd)
		config.Header = req.Header
	}

	return config, nil
}

func wsDial(url *url.URL, user, passwd string) {
	cfg, err := newConfig(url, user, passwd)
	if err != nil {
		LogError("Configuring websocket", "err", err)
		return
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(cfg)
		if err == nil {
			// Service the client websocket
			wsClient(conn)
		} else {
			LogError("Dialing", "url", url, "err", err)
		}

		// Try again in a second
		time.Sleep(time.Second)
	}
}

func wsClient(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{conn: conn}
	var pkt = &Packet{
		Dst:  root.Id,
		Path: "/announce",
	}

	devicesMu.RLock()
	pkt.Marshal(aliveDevices())
	devicesMu.RUnlock()

	// Send announcement
	LogInfo("<- Sending", "pkt", pkt)
	err := link.Send(pkt)
	if err != nil {
		LogError("Sending", "err", err)
		return
	}

	// Receive welcome within 1 sec
	pkt, err = link.receiveTimeout(time.Second)
	if err != nil {
		LogError("Receiving", "err", err)
		return
	}

	LogInfo("-> Reply", "pkt", pkt)
	if pkt.Path != "/welcome" {
		LogError("Not welcomed, got", "path", pkt.Path)
		return
	}

	LogInfo("Adding Uplink")
	uplinksAdd(link)

	// Send /online packet to all online devices
	devicesOnline(link)

	// Route incoming packets down to the destination device.  Stop and
	// disconnect on EOF.

	LogInfo("Receiving packets")
	for {
		pkt, err := link.receivePoll()
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		LogDebug("-> Route packet DOWN", "pkt", pkt)
		downlinksRoute(pkt)
	}

	LogInfo("Removing Uplink")
	uplinksRemove(link)
}
