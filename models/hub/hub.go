package hub

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"html/template"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
)

//go:embed css html images js
var fs embed.FS
var tmpls = template.Must(template.ParseFS(fs, "html/*"))

type Device struct {
	Model string
	Name  string
	Online  bool
	thinger dean.Thinger
}

type Devices map[string]*Device        // keyed by id
type Models []string
type makers map[string]dean.ThingMaker // keyed by model

type Hub struct {
	*common.Common
	Devices
	Models
	makers   makers
}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	return &Hub{
		Common: common.New(id, model, name).(*common.Common),
		/*
		Devices:  Devices{
			"88_ae_dd_0a_70_92": &Device{Model: "relays", Name: "Relays"},
			"d1":                &Device{Model: "relays", Name: "Relays01"},
			"d2":                &Device{Model: "relays", Name: "Relays02"},
		},
		*/
		makers:   makers{},
	}
}

func (h *Hub) Register(model string, maker dean.ThingMaker) {
	h.makers[model] = maker
}

func (h *Hub) Unregister(model string) {
	delete(h.makers, model)
}

func (h *Hub) Make(id, model, name string) dean.Thinger {
	dev := h.Devices[id]
	if dev == nil {
		return nil
	}
	return dev.thinger
}

func (h *Hub) getState(msg *dean.Msg) {
	h.Path = "state"
	h.Models = make([]string, 0, len(h.makers))
	for model := range h.makers {
		h.Models = append(h.Models, model)
	}
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

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":    h.getState,
		"connected":    h.connect(true),
		"disconnected": h.connect(false),
	}
}

func (h *Hub) storeDevices() {
	bytes, _ := json.MarshalIndent(h.Devices, "", "\t")
	os.WriteFile("devices.json", bytes, 0600)
}

func (h *Hub) makeThingers() {
	for id, dev := range h.Devices {
		dev.Online = false
		if maker, ok := h.makers[dev.Model]; ok {
			dev.thinger = maker(id, dev.Model, dev.Name)
		}
	}
}

func (h *Hub) restoreDevices() {
	bytes, err := os.ReadFile("devices.json")
	if err == nil {
		json.Unmarshal(bytes, &h.Devices)
	} else {
		println(err.Error())
	}
}

func (h *Hub) dumpDevices() {
	b, _ := json.MarshalIndent(h.Devices, "", "\t")
	fmt.Println(string(b))
}

func (h *Hub) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("api\n"))
	w.Write([]byte("create?id&model&name\n"))
	w.Write([]byte("deploy?id\n"))
}

func (h *Hub) create(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")

	println("create", id, model, name)

	if !dean.ValidId(id) || !dean.ValidId(model) || !dean.ValidId(name) {
		http.Error(w, "Invalid id|model|name", http.StatusNotAcceptable)
		return
	}

	dev := h.Devices[id]
	if dev != nil {
		http.Error(w, "Device already exists", http.StatusNotAcceptable)
		return
	}

	maker, ok := h.makers[model]
	if !ok {
		http.Error(w, "Device model unknown", http.StatusNotAcceptable)
		return
	}

	h.Devices[id] = &Device{model, name, false, maker(id, model, name)}
	h.storeDevices()
}

func (h *Hub) deploy(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	println("deploy", id)
}

func (h *Hub) API(fs embed.FS, tmpls *template.Template, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api":
		h.api(w, r)
	case "/create":
		h.create(w, r)
	case "/deploy":
		h.deploy(w, r)
	default:
		h.Common.API(fs, tmpls, w, r)
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.API(fs, tmpls, w, r)
}

func (h *Hub) Run(i *dean.Injector) {
	//h.storeDevices()
	h.restoreDevices()
	h.makeThingers()
	h.dumpDevices()
	select {}
}
