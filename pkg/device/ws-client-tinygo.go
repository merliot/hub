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

func (s *server) wsDial(url *url.URL, user, passwd string) {
	cfg, err := newConfig(url, user, passwd)
	if err != nil {
		s.logError("Configuring websocket", "err", err)
		return
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(cfg)
		if err == nil {
			// Service the client websocket
			s.wsClient(url.String(), conn)
		} else {
			s.logError("Dialing", "url", url, "err", err)
		}

		// Try again in a second
		time.Sleep(time.Second)
	}
}

// TODO: merge this with ws-client.go:wsClient() and share?  Or maybe just
// parts of it?
func (s *server) wsClient(url string, conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{name: "/ws client dialing " + url, conn: conn}
	var pkt = &Packet{
		Dst:  s.root.Id,
		Path: "announce",
	}

	devices := make(map[string]*device)
	devices[s.root.Id] = s.root

	pkt.Marshal(devices)

	// Send announcement
	s.logInfo("<- Sending", "pkt", pkt)
	err := link.Send(pkt)
	if err != nil {
		s.logError("Sending", "err", err)
		return
	}

	// Receive welcome within 1 sec
	pkt, err = s.receiveTimeout(link, time.Second)
	if err != nil {
		s.logError("Receiving", "err", err)
		return
	}

	s.logInfo("-> Reply", "pkt", pkt)
	if pkt.Path != "welcome" {
		s.logError("Not welcomed, got", "path", pkt.Path)
		return
	}

	s.logInfo("Adding Uplink")
	s.uplinks.add(link)

	// Send online packet for all online devices
	s.devicesOnline(link)

	// Route incoming packets down to the destination device
	s.logInfo("Receiving packets")
	for {
		pkt, err := s.receive(link)
		if err != nil {
			s.logError("Receiving packet", "err", err)
			break
		}
		s.logDebug("-> Route packet DOWN", "pkt", pkt)
		if err := s.routeDown(pkt); err != nil {
			s.logError("Routing packet DOWN", "err", err)
			break
		}
	}

	s.logInfo("Removing Uplink")
	s.uplinks.remove(link)
}
