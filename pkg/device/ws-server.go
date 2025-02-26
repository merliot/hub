//go:build !tinygo

package device

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins, update as necessary
		return true
	},
}

func (s *server) wsHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		LogError("Upgrading WebSocket", "err", err)
		return
	}
	s.wsServer(conn)
}

func (d *device) handleAnnounced(pkt *Packet) {
	if _, err := handleAnnounce(pkt); err != nil {
		LogDebug("Error handling announcement", "err", err)
	}
}

func handleAnnounce(pkt *Packet) (id string, err error) {

	var annDevices = make(deviceMap)

	pkt.Unmarshal(&annDevices)

	// Find root device in announcement

	annRoot, err := findRoot(annDevices)
	if err != nil {
		return "", fmt.Errorf("Cannot find root: %s", err)
	}
	id = annRoot.Id

	if id != pkt.Dst {
		return "", fmt.Errorf("Id mismatch announcement-id: %s pkt-id: %s", id, pkt.Dst)
	}

	if id == root.Id {
		return "", fmt.Errorf("Cannot dial into root (self)")
	}

	// Make sure announcement root device exists in existing devices and
	// matches existing device

	if err := validate(annRoot); err != nil {
		return "", fmt.Errorf("Announcement mismatch id: %s err: %s", id, err)
	}

	// Merge in annoucement devices

	if err := merge(devices, annDevices); err != nil {
		return "", fmt.Errorf("Cannot merge device id: %s err: %s", id, err)
	}

	// Rebuild routing table
	routesBuild(root)

	// Send /announced packet up so parents can update their trees
	pkt.SetPath("/announced")
	LogDebug("<- Sending", "pkt", pkt)
	pkt.RouteUp()

	return id, nil
}

func (s *server) wsServer(conn *websocket.Conn) {

	defer conn.Close()

	var link = &wsLink{conn: conn}

	// First receive should be an /announce packet
	pkt, err := link.receive()
	if err != nil {
		LogError("Receiving first packet", "err", err)
		return
	}

	if pkt.Path != "/announce" {
		LogError("Expected /announce, got", "path", pkt.Path)
		return
	}
	LogDebug("-> Announcement", "pkt", pkt)

	id, err := handleAnnounce(pkt)
	if err != nil {
		LogError("Bad announcement", "err", err)
		return
	}

	// Announcement is good, send /welcome packet down to device
	pkt.ClearMsg().SetPath("/welcome")
	LogDebug("<- Sending", "pkt", pkt)
	link.Send(pkt)

	// Add as active downlink
	LogDebug("Adding Downlink", "id", id)
	s.downlinks.add(id, link)

	// Start ping/pong
	link.setPongHandler()
	link.startPing()

	// Route incoming packets up to the destination device
	for {
		pkt, err := link.receive()
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}

		// Special handling for non-gorilla clients
		// TODO: delete this when clients are converted to gorilla websocket
		if pkt.Path == "/ping" {
			link.Send(pkt.SetPath("/pong"))
			continue
		}

		LogDebug("-> Received", "pkt", pkt)
		if err := s.handle(pkt); err != nil {
			LogError("Handling packet", "err", err)
		}
	}

	LogDebug("Removing Downlink", "id", id)
	s.downlinks.remove(id)

	s.deviceOffline(id)
}
