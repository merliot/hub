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
			s.logError("Dialing", "url", wsURL, "err", err)
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
			s.wsClient(wsURL.String(), conn)
		} else {
			s.logError("Dialing", "url", wsURL, "err", err)
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

		pkt := d.newPacket()
		pkt.SetPath("online").Marshal(d.State)

		s.logInfo("Sending", "pkt", pkt)
		l.Send(pkt)

		return true
	})
}

func (s *server) wsClient(dialURL string, conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{name: "/ws client dialing " + dialURL, conn: conn}
	var pkt = s.root.newPacket()

	link.setPongHandler()
	link.startPing()

	pkt.SetPath("announce").Marshal(s.devices.getJSON())

	// Send announcement
	s.logInfo("<- Sending", "pkt", pkt)
	err := link.Send(pkt)
	if err != nil {
		s.logError("Sending", "err", err)
		return
	}

	// Receive welcome
	pkt, err = s.receive(link)
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

	// Send online packet to all online devices
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
