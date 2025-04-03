//go:build !tinygo

package device

import (
	"net/http"
)

func (s *server) packetHandlersInstall(d *device) {
	for path, handler := range d.PacketHandlers {
		if len(path) > 0 && path[0] != '/' {
			s.LogError("Packet handler missing leading '/'", "path", path, "device", d)
			continue
		}
		d.Handle("POST "+path, s.newPacketRoute(handler, d))
	}
}

func (s *server) newPacketRoute(h packetHandler, d *device) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionId := r.Header.Get("session-id")

		if s.sessions.expired(sessionId) {
			// Force full page refresh to start new session
			s.LogDebug("Session expired, refreshing", "id", sessionId)
			w.Header().Set("HX-Refresh", "true")
			return
		}

		msg := h.gen()
		pkt, err := s.newPacketFromRequest(r, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.Id)

		if d.isSet(flagMetal) {
			d.handle(pkt)
		} else {
			if err := s.routeDown(pkt); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	})
}
