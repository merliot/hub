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

func (s *server) handleAnnounced(pkt *Packet) {
	if _, err := s.handleAnnounce(pkt); err != nil {
		LogDebug("Error handling announcement", "err", err)
	}
}

func (s *server) match(a *device) error {

	d, exists := s.devices.get(a.Id)
	if !exists {
		return deviceNotFound(a.Id)
	}

	if d.Model != a.Model {
		return fmt.Errorf("Device model wrong.  Want %s; got %s",
			d.Model, a.Model)
	}

	if d.Name != a.Name {
		return fmt.Errorf("Device name wrong.  Want %s; got %s",
			d.Name, a.Name)
	}

	if d.DeployParams != a.DeployParams {
		return fmt.Errorf("Device DeployParams wrong.\nWant: %s\nGot: %s",
			d.DeployParams, a.DeployParams)
	}

	return nil
}

func (s *server) handleAnnounce(pkt *Packet) (id string, err error) {

	var annDevicesJSON devicesJSON
	var annDevices deviceMap

	pkt.Unmarshal(&annDevicesJSON)
	annDevices.loadJSON(annDevicesJSON)

	// Find root device in announcement

	anchor, err := annDevices.buildTree()
	if err != nil {
		return "", fmt.Errorf("Cannot find root: %s", err)
	}
	id = anchor.Id

	if id != pkt.Dst {
		return "", fmt.Errorf("Id mismatch announcement-id: %s pkt-id: %s", id, pkt.Dst)
	}

	if id == s.root.Id {
		return "", fmt.Errorf("Cannot dial into root (self)")
	}

	// Make sure announcement root device exists in existing devices and
	// matches existing device

	if err := s.match(anchor); err != nil {
		return "", fmt.Errorf("Announcement mismatch id: %s err: %s", id, err)
	}

	// Merge in annoucement devices

	if err := s.merge(id, annDevices); err != nil {
		return "", fmt.Errorf("Cannot merge device id: %s err: %s", id, err)
	}

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
	pkt, err := s.receive(link)
	if err != nil {
		LogError("Receiving first packet", "err", err)
		return
	}
	pkt.server = s

	if pkt.Path != "/announce" {
		LogError("Expected /announce, got", "path", pkt.Path)
		return
	}
	LogDebug("-> Announcement", "pkt", pkt)

	id, err := s.handleAnnounce(pkt)
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
		pkt, err := s.receive(link)
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		pkt.server = s

		// Special handling for non-gorilla clients
		// TODO: delete this when clients are converted to gorilla websocket
		if pkt.Path == "/ping" {
			link.Send(pkt.SetPath("/pong"))
			continue
		}

		LogDebug("-> Received", "pkt", pkt)
		if err := pkt.handle(); err != nil {
			LogError("Handling packet", "err", err)
		}
	}

	LogDebug("Removing Downlink", "id", id)
	s.downlinks.remove(id)

	s.deviceOffline(id)
}
