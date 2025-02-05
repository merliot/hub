package device

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func wsDial(wsURL *url.URL, user, passwd string) {

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
			wsClient(conn)
		} else {
			LogError("Dialing", "url", wsURL, "err", err)
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

	link.setPongHandler()
	link.startPing()

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

	// Receive welcome
	pkt, err = link.receive()
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

	// Route incoming packets down to the destination device
	LogInfo("Receiving packets")
	for {
		pkt, err := link.receive()
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
