//go:build !tinygo

package device

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func (s *server) wsHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logError("Upgrading WebSocket", "err", err)
		return
	}
	s.wsServer(conn)
}

func (s *server) handleAnnounced(pkt *Packet) {
	if _, err := s.handleAnnounce(pkt); err != nil {
		s.logDebug("Error handling announcement", "err", err)
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

	// Hold lock while merging new devices in with existing
	// devices...we could have multiple announcements coming in on
	// different requests, so want to serialize the merging.

	s.Lock()
	defer s.Unlock()

	// Make sure announcement root device exists in existing devices and
	// matches existing device

	if err := s.match(anchor); err != nil {
		return "", fmt.Errorf("Announcement mismatch id: %s err: %s", id, err)
	}

	// Merge in annoucement devices

	if err := s.merge(id, annDevices); err != nil {
		return "", fmt.Errorf("Cannot merge device id: %s err: %s", id, err)
	}

	// Send announced packet up so parents can update their trees
	pkt.SetPath("announced")
	s.logDebug("<- Sending", "pkt", pkt)
	pkt.RouteUp()

	return id, nil
}

func (s *server) wsServer(conn *websocket.Conn) {

	defer conn.Close()

	var link = &wsLink{conn: conn}

	// First receive should be an announce packet
	pkt, err := s.receive(link)
	if err != nil {
		s.logError("Receiving first packet", "err", err)
		return
	}
	pkt.server = s

	if pkt.Path != "announce" {
		s.logError("Expected announce, got", "path", pkt.Path)
		return
	}
	s.logDebug("-> Announcement", "pkt", pkt)

	id, err := s.handleAnnounce(pkt)
	if err != nil {
		s.logError("Bad announcement", "err", err)
		return
	}

	// Announcement is good, send welcome packet down to device
	pkt.ClearMsg().SetPath("welcome")
	s.logDebug("<- Sending", "pkt", pkt)
	link.Send(pkt)

	// Add as active downlink
	s.logDebug("Adding Downlink", "id", id)
	s.downlinks.add(id, link)

	// Start ping/pong
	link.setPongHandler()
	link.startPing()

	// Route incoming packets up to the destination device
	for {
		pkt, err := s.receive(link)
		if err != nil {
			s.logError("Receiving packet", "err", err)
			break
		}
		pkt.server = s

		// Special handling for non-gorilla clients
		// TODO: delete this when clients are converted to gorilla websocket
		if pkt.Path == "ping" {
			link.Send(pkt.SetPath("pong"))
			continue
		}

		s.logDebug("-> Received", "pkt", pkt)
		if err := pkt.handle(); err != nil {
			s.logError("Handling packet", "err", err)
		}
	}

	s.logDebug("Removing Downlink", "id", id)
	s.downlinks.remove(id)

	s.deviceOffline(id)
}
