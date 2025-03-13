//go:build !tinygo

package device

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func (s *server) wsDial(wsURL *url.URL, user, passwd string) {

	var hdr = http.Header{}

	// If valid user, set the basic auth header for the request
	if user != "" {
		req, err := http.NewRequest("GET", wsURL.String(), nil)
		if err != nil {
			LogError("Dialing", "url", wsURL, "err", err)
			return
		}
		req.SetBasicAuth(user, passwd)
		hdr = req.Header
	}

	for {
		// Connect to the server with custom headers
		conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), hdr)
		if err == nil {
			// Service the client websocket
			s.wsClient(conn)
		} else {
			LogError("Dialing", "url", wsURL, "err", err)
		}

		// Try again in a second
		time.Sleep(time.Second)
	}
}

func (s *server) devicesOnline(l linker) {

	s.devices.drange(func(id string, d *device) bool {

		if !d.isSet(flagOnline) {
			return true
		}

		pkt := s.newPacket()
		pkt.SetDst(id).SetPath("/online").Marshal(d.State)

		LogInfo("Sending", "pkt", pkt)
		l.Send(pkt)

		return true
	})
}

func (s *server) wsClient(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{conn: conn}
	var pkt = s.newPacket()

	link.setPongHandler()
	link.startPing()

	pkt.SetPath("/announce").Marshal(s.devices.getJSON())

	// Send announcement
	LogInfo("<- Sending", "pkt", pkt)
	err := link.Send(pkt)
	if err != nil {
		LogError("Sending", "err", err)
		return
	}

	// Receive welcome
	pkt, err = s.receive(link)
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
	s.uplinks.add(link)

	// Send /online packet to all online devices
	s.devicesOnline(link)

	// Route incoming packets down to the destination device
	LogInfo("Receiving packets")
	for {
		pkt, err := s.receive(link)
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		LogDebug("-> Route packet DOWN", "pkt", pkt)
		if err := s.routeDown(pkt); err != nil {
			LogError("Routing packet DOWN", "err", err)
			break
		}
	}

	LogInfo("Removing Uplink")
	s.uplinks.remove(link)
}
