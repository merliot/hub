//go:build !tinygo

package hub

import (
	"net/http"
)

func (d *device) handlersInstall() {
	for path, handler := range d.Handlers {
		if len(path) > 0 && path[0] != '/' {
			LogError("Handler missing leading '/'", "path", path, "device", d)
			continue
		}
		d.Handle("POST "+path, d.newPacketRoute(handler))
	}
}

func (d *device) newPacketRoute(h handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := h.gen()
		pkt, err := newPacketFromURL(r.URL, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.Id)
		if d.IsSet(flagMetal) {
			d.handle(pkt)
		} else {
			pkt.RouteDown()
		}
	})
}
