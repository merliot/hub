//go:build !tinygo

package device

import (
	"net/http"
)

func (d *device) packetHandlersInstall() {
	for path, handler := range d.PacketHandlers {
		if len(path) > 0 && path[0] != '/' {
			LogError("Packet handler missing leading '/'", "path", path, "device", d)
			continue
		}
		d.Handle("POST "+path, d.newPacketRoute(handler))
	}
}

func (d *device) newPacketRoute(h packetHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionId := r.Header.Get("session-id")

		if sessionExpired(sessionId) {
			// Force full page refresh to start new session
			LogDebug("Session expired, refreshing", "id", sessionId)
			w.Header().Set("HX-Refresh", "true")
			return
		}

		msg := h.gen()
		pkt, err := newPacketFromRequest(r, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.Id)

		if d.isSet(flagMetal) {
			d.handle(pkt)
		} else {
			pkt.RouteDown()
		}
	})
}
