//go:build !tinygo

package device

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// ws handles /ws requests
func wsHandle(w http.ResponseWriter, r *http.Request) {
	serv := websocket.Server{Handler: websocket.Handler(wsServer)}
	serv.ServeHTTP(w, r)
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
	LogDebug("Announcement", "pkt", pkt)

	var annDevices = make(deviceMap)
	pkt.Unmarshal(&annDevices)

	// Find root device in announcement

	annRoot, err := findRoot(annDevices)
	if err != nil {
		LogDebug("Cannot find root", "err", err)
		return
	}
	id := annRoot.Id

	if id != pkt.Dst {
		LogDebug("Id mismatch", "announcement-id", id, "pkt-id", pkt.Dst)
		return
	}

	if id == root.Id {
		LogDebug("Cannot dial into root (self)", "id", id)
		return
	}

	// Make sure announcement root device exists in existing devices and
	// matches existing device

	if err := validate(annRoot); err != nil {
		LogDebug("Announcement mismatch", "id", id, "err", err)
		return
	}

	// Merge in annoucement devices

	if err := merge(devices, annDevices); err != nil {
		LogDebug("Cannot merge device", "id", id, "err", err)
		return
	}

	// Rebuild routing table
	routesBuild(root)

	// Announcement is good, reply with /welcome packet

	pkt.ClearMsg().SetPath("/welcome")
	LogDebug("Sending welcome", "pkt", pkt)
	link.Send(pkt)

	// Add as active download link

	LogInfo("Adding Downlink", "id", id)
	downlinksAdd(id, link)

	// Route incoming packets up to the destination device.  Stop and
	// disconnect on EOF.

	for {
		pkt, err := link.receivePoll()
		if err != nil {
			LogError("Receiving packet", "err", err)
			break
		}
		LogDebug("Route packet UP", "pkt", pkt)
		deviceRouteUp(pkt.Dst, pkt)
	}

	LogInfo("Removing Downlink", "id", id)
	downlinksRemove(id)

	deviceOffline(id)
}
