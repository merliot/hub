//go:build !tinygo

package device

import (
	"net/http"
	"strings"
)

func (s *server) packetHandlersInstall(d *device) {
	for path, handler := range d.PacketHandlers {
		if strings.HasPrefix(path, "/") {
			d.Handle("POST "+path, s.newPacketRoute(handler, d))
		}
	}
}

func (s *server) newPacketRoute(h packetHandler, d *device) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
