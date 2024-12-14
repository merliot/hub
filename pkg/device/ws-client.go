//go:build !tinygo

// TODO get gorilla/websocket working on tinygo.  Currently hit:
//       ../../../go/pkg/mod/github.com/gorilla/websocket@v1.5.1/client.go:18:2: package net/http/httptrace is not in std (/root/...

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

	link.setPongHandler()
	link.startPing()

	pkt.Marshal(&ann)

	// Send announcement
	LogInfo("Sending announcement", "pkt", pkt)
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
		pkt, err := link.receive()
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		LogDebug("Route packet DOWN", "pkt", pkt)
		deviceRouteDown(pkt.Dst, pkt)
	}

	LogInfo("Removing Uplink")
	uplinksRemove(link)
}
