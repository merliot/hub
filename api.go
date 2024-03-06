package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/merliot/device"
)

func (h *Hub) apiCreate(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")

	thinger, err := h.server.CreateThing(id, model, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	child := thinger.(device.Devicer)
	child.CopyWifiAuth(h.WifiAuth)
	child.SetWsScheme(h.WsScheme)
	child.Load(filePath(id))

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Child id '%s' created", id)
}

func (h *Hub) apiDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h.server.DeleteThing(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Child id '%s' deleted", id)
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "create":
		h.apiCreate(w, r)
	case "delete":
		h.apiDelete(w, r)
	case "devices":
		data, _ := json.MarshalIndent(h.Children, "", "\t")
		h.RenderTemplate(w, "devices.tmpl", string(data))
	case "models":
		h.RenderTemplate(w, "models.tmpl", h)
	default:
		h.API(w, r, h)
	}
}
