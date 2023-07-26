package hub

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
)

//go:embed css html images js
var fs embed.FS
var tmpls = template.Must(template.ParseFS(fs, "html/*"))

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
	server *dean.Server
	async  chan *dean.Msg
}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	return &Hub{
		Common:  common.New(id, model, name).(*common.Common),
		Devices: make(Devices),
		async:   make(chan *dean.Msg, 10),
	}
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

func (h *Hub) broadcast(msg *dean.Msg) {
	msg.Broadcast()
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":     h.getState,
		"connected":     h.connect(true),
		"disconnected":  h.connect(false),
		"device/create": h.broadcast,
		"device/delete": h.broadcast,
	}
}

type DeviceCreateMsg struct {
	Path  string
	Id    string
	Model string
	Name  string
}

func (h *Hub) createDevice(id, model, name string) error {
	var msg dean.Msg
	err := h.server.CreateThing(id, model, name)
	if err == nil {
		h.Devices[id] = &Device{Model: model, Name: name}
		h.storeDevices()
		h.async <- msg.Marshal(&DeviceCreateMsg{"device/create", id, model, name})
	}
	return err
}

type DeviceDeleteMsg struct {
	Path string
	Id   string
}

func (h *Hub) deleteDevice(id string) error {
	var msg dean.Msg
	err := h.server.DeleteThing(id)
	if err == nil {
		delete(h.Devices, id)
		h.storeDevices()
		h.async <- msg.Marshal(&DeviceDeleteMsg{"device/delete", id})
	}
	return err
}

func (h *Hub) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/create?id={id}&model={model}&name={name}\n"))
	w.Write([]byte("/delete?id={id}\n"))
	w.Write([]byte("/deploy?id={id}\n"))
}

func (h *Hub) apiCreate(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")
	err := h.createDevice(id, model, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Hub) apiDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h.deleteDevice(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Hub) _deploy(id string) error {
	_, ok := h.Devices[id]
	if !ok {
		return fmt.Errorf("Device ID '%s' doesn't exist!", id)
	}
	// TODO build binary and download
	return nil
}

func (h *Hub) apiDeploy(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h._deploy(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Hub) API(fs embed.FS, tmpls *template.Template, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api":
		h.api(w, r)
	case "/create":
		h.apiCreate(w, r)
	case "/delete":
		h.apiDelete(w, r)
	case "/deploy":
		h.apiDeploy(w, r)
	default:
		h.Common.API(fs, tmpls, w, r)
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.API(fs, tmpls, w, r)
}

func (h *Hub) restoreDevices() {
	var devices Devices
	bytes, _ := os.ReadFile("devices.json")
	json.Unmarshal(bytes, &devices)
	for id, dev := range devices {
		err := h.createDevice(id, dev.Model, dev.Name)
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

	for {
		select {
		case msg := <-h.async:
			i.Inject(msg)
		}
	}
}
