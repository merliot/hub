package hub

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
)

//go:embed css images js template
var fs embed.FS

type Device struct {
	Model  string
	Name   string
	Online bool
}

type Devices map[string]*Device // keyed by id
type Models []string

type Hub struct {
	*common.Common
	Devices
	Models
	server    *dean.Server
	templates *template.Template
}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	h := &Hub{}
	h.Common = common.New(id, model, name).(*common.Common)
	h.Devices = make(Devices)
	h.CompositeFs.AddFS(fs)
	h.templates = h.CompositeFs.ParseFS("template/*")
	return h
}

func (h *Hub) UseServer(server *dean.Server) {
	h.server = server
}

func (h *Hub) getState(msg *dean.Msg) {
	h.Path = "state"
	msg.Marshal(h).Reply()
}

func (h *Hub) online(msg *dean.Msg, online bool) {
	var thing dean.ThingMsgConnect
	msg.Unmarshal(&thing)

	if dev, ok := h.Devices[thing.Id]; ok {
		dev.Online = online
	}

	msg.Broadcast()
}

func (h *Hub) connect(online bool) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		h.online(msg, online)
	}
}

func (h *Hub) createdThing(msg *dean.Msg) {
	var create dean.ThingMsgCreated
	msg.Unmarshal(&create)
	h.Devices[create.Id] = &Device{Model: create.Model, Name: create.Name}
	h.storeDevices()
	create.Path = "created/device"
	msg.Marshal(&create).Broadcast()
}

func (h *Hub) deletedThing(msg *dean.Msg) {
	var del dean.ThingMsgDeleted
	msg.Unmarshal(&del)
	delete(h.Devices, del.Id)
	h.storeDevices()
	del.Path = "deleted/device"
	msg.Marshal(&del).Broadcast()
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":     h.getState,
		"connected":     h.connect(true),
		"disconnected":  h.connect(false),
		"created/thing": h.createdThing,
		"deleted/thing": h.deletedThing,
	}
}

func (h *Hub) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/create?id={id}&model={model}&name={name}\n"))
	w.Write([]byte("/delete?id={id}\n"))
}

func (h *Hub) apiCreate(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")
	err := h.server.CreateThing(id, model, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Device id '%s' created", id)
}

func (h *Hub) apiDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h.server.DeleteThing(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Device id '%s' deleted", id)
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "api":
		h.api(w, r)
	case "create":
		h.apiCreate(w, r)
	case "delete":
		h.apiDelete(w, r)
	default:
		h.API(h.templates, w, r)
	}
}

func (h *Hub) restoreDevices() {
	var devices Devices
	bytes, _ := os.ReadFile("devices.json")
	json.Unmarshal(bytes, &devices)
	for id, dev := range devices {
		err := h.server.CreateThing(id, dev.Model, dev.Name)
		if err != nil {
			fmt.Printf("Error creating device Id '%s': %s\n", id, err)
		}
	}
}

func (h *Hub) storeDevices() {
	bytes, _ := json.MarshalIndent(h.Devices, "", "\t")
	os.WriteFile("devices.json", bytes, 0600)
}

func (h *Hub) dumpDevices() {
	b, _ := json.MarshalIndent(h.Devices, "", "\t")
	fmt.Println(string(b))
}

func (h *Hub) Run(i *dean.Injector) {
	h.Models = h.server.GetModels()
	h.restoreDevices()
	select {}
}
