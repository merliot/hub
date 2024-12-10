package device

import (
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func wsDial(wsURL *url.URL, user, passwd string) {
	for {
		// Connect to the server
		conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
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
	var ann = announcement{
		Id:           root.Id,
		Model:        root.Model,
		Name:         root.Name,
		DeployParams: root.DeployParams,
	}
	var pkt = &Packet{
		Dst:  ann.Id,
		Path: "/announce",
	}

	pkt.Marshal(&ann)

	// Send announcement
	LogInfo("Sending announcement", "pkt", pkt)
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

	LogInfo("Reply from announcement", "pkt", pkt)
	if pkt.Path != "/welcome" {
		LogError("Not welcomed, got", "path", pkt.Path)
		return
	}

	LogInfo("Adding Uplink")
	uplinksAdd(link)

	// Send /state packets to all devices
	devicesSendState(link)

	// Route incoming packets down to the destination device
	LogInfo("Receiving packets")
	for {
		pkt, err := link.receivePoll()
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		LogInfo("Route packet DOWN", "pkt", pkt)
		deviceRouteDown(pkt.Dst, pkt)
	}

	LogInfo("Removing Uplink")
	uplinksRemove(link)
}
