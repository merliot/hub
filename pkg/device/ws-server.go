//go:build !tinygo

package device

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins, update as necessary
		return true
	},
}

func wsHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		LogError("Upgrading WebSocket", "err", err)
		return
	}
	wsServer(conn)
}

func wsServer(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{conn: conn}

	// First receive should be an /announce packet
	pkt, err := link.receive()
	if err != nil {
		LogError("Receiving first packet", "err", err)
		return
	}

	if pkt.Path != "/announce" {
		LogError("Expected announcement, got", "path", pkt.Path)
		return
	}
	LogInfo("Announcement", "pkt", pkt)

	var ann announcement
	pkt.Unmarshal(&ann)

	if ann.Id != pkt.Dst {
		LogError("Id mismatch", "announcement-id", ann.Id, "pkt-id", pkt.Dst)
		return
	}

	if ann.Id == root.Id {
		LogError("Cannot dial into root (self)", "id", ann.Id)
		return
	}

	if err := deviceOnline(ann); err != nil {
		LogError("Cannot switch device online", "id", ann.Id, "err", err)
		return
	}

	// Announcement is good, reply with /welcome packet
	pkt.SetPath("/welcome")
	LogInfo("Sending welcome", "pkt", pkt)
	link.Send(pkt)

	// Add as active downlink
	id := ann.Id
	LogInfo("Adding Downlink", "id", ann.Id)
	downlinksAdd(id, link)

	// Route incoming packets up to the destination device
	for {
		pkt, err := link.receivePoll()
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		LogInfo("Route packet UP", "pkt", pkt)
		deviceRouteUp(pkt.Dst, pkt)
	}

	LogInfo("Removing Downlink", "id", ann.Id)
	downlinksRemove(id)

	deviceOffline(id)
}
