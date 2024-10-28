//go:build !tinygo

package hub

import (
	"log/slog"
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
		slog.Error("Receiving first packet", "err", err)
		return
	}

	if pkt.Path != "/announce" {
		slog.Error("Expected announcement, got", "path", pkt.Path)
		return
	}
	slog.Info("Announcement", "pkt", pkt)

	var ann announcement
	pkt.Unmarshal(&ann)

	if ann.Id != pkt.Dst {
		slog.Error("Id mismatch", "announcement-id", ann.Id, "pkt-id", pkt.Dst)
		return
	}

	if ann.Id == root.Id {
		slog.Error("Cannot dial into root (self)", "id", ann.Id)
		return
	}

	if err := deviceOnline(ann); err != nil {
		slog.Error("Cannot switch device online", "id", ann.Id, "err", err)
		return
	}

	// Announcement is good, reply with /welcome packet

	pkt.SetPath("/welcome")
	slog.Info("Sending welcome", "pkt", pkt)
	link.Send(pkt)

	// Add as active download link

	id := ann.Id
	slog.Info("Adding Downlink", "id", ann.Id)
	downlinksAdd(id, link)

	// Route incoming packets up to the destination device.  Stop and
	// disconnect on EOF.

	for {
		pkt, err := link.receivePoll()
		if err != nil {
			slog.Error("Receiving packet", "err", err)
			break
		}
		slog.Info("Route packet UP", "pkt", pkt)
		deviceRouteUp(pkt.Dst, pkt)
	}

	slog.Info("Removing Downlink", "id", ann.Id)
	downlinksRemove(id)

	deviceOffline(id)
}
