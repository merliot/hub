package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func (h *Hub) apiCreate(w http.ResponseWriter, r *http.Request) {
	if h.Locked {
		http.Error(w, "Refusing to create device, hub is locked", http.StatusLocked)
		return
	}

	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")

	thinger, err := h.server.CreateThing(id, model, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.loadDevice(thinger, id, "")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Child id '%s' created\n", id)
}

func (h *Hub) apiDelete(w http.ResponseWriter, r *http.Request) {
	if h.Locked {
		http.Error(w, "Refusing to delete device, hub is locked", http.StatusLocked)
		return
	}
	id := r.URL.Query().Get("id")
	err := h.server.DeleteThing(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Child id '%s' deleted\n", id)
}

func (h *Hub) apiDevices(w http.ResponseWriter, r *http.Request) {
	devices := make(Children)
	for id, child := range h.Children {
		child.DeployParams = child.Devicer.GetDeployParams()
		devices[id] = child
	}
	data, _ := json.MarshalIndent(devices, "", "\t")
	w.Write(data)
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "create":
		h.apiCreate(w, r)
	case "delete":
		h.apiDelete(w, r)
	case "devices":
		h.apiDevices(w, r)
	case "models":
		h.RenderTemplate(w, "models.tmpl", h)
	default:
		h.API(w, r, h)
	}
}

func (h *Hub) ModelsInUse() map[string]bool {
	models := make(map[string]bool)
	for _, child := range h.Children {
		models[child.Model] = true
	}
	return models
}

func (h *Hub) ChildrenSorted() []string {
	keys := make([]string, 0, len(h.Children))
	for key := range h.Children {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return h.Children[keys[i]].Name < h.Children[keys[j]].Name
	})
	return keys
}
