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
	makers  makers
	async chan *dean.Msg
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
		async: make(chan *dean.Msg),
	}
}

func (h *Hub) Register(model string, maker dean.ThingMaker) {
	h.makers[model] = maker
}

func (h *Hub) Unregister(model string) {
	delete(h.makers, model)
}

func (h *Hub) Make(id, model, name string) dean.Thinger {
	// Want exact match on [id, model, name]
	dev := h.Devices[id]
	if dev == nil {
		return nil
	}
	if dev.Model != model || dev.Name != name {
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

type MsgCreate struct {
	dean.ThingMsg
	Id string
	Model string
	Name string
	Err string
}

func (h *Hub) create(msg *dean.Msg) {
	var m MsgCreate
	msg.Unmarshal(&m)

	err := h._create(m.Id, m.Model, m.Name)
	if err == nil {
		m.Path = "create/good"
		m.Err = ""
	} else {
		m.Path = "create/bad"
		m.Err = err.Error()
	}

	msg.Marshal(&m).Reply().Broadcast()
}

type MsgDelete struct {
	dean.ThingMsg
	Id string
	Err string
}

func (h *Hub) deletef(msg *dean.Msg) {
	var m MsgDelete
	msg.Unmarshal(&m)

	err := h._delete(m.Id)
	if err == nil {
		m.Path = "delete/good"
		m.Err = ""
	} else {
		m.Path = "delete/bad"
		m.Err = err.Error()
	}

	msg.Marshal(&m).Reply().Broadcast()
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":    h.getState,
		"connected":    h.connect(true),
		"disconnected": h.connect(false),
		"create":       h.create,
		"delete":       h.deletef,
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
	w.Write([]byte("/api\n"))
	w.Write([]byte("/create?id={id}&model={model}&name={name}\n"))
	w.Write([]byte("/delete?id={id}\n"))
	w.Write([]byte("/deploy?id={id}\n"))
}

func (h *Hub) _create(id, model, name string)  error {
	if !dean.ValidId(id) {
		return fmt.Errorf("Invalid ID.  A valid ID is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}
	if !dean.ValidId(model) {
		return fmt.Errorf("Invalid Model.  A valid Model is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}
	if !dean.ValidId(name) {
		return fmt.Errorf("Invalid Name.  A valid Name is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}

	dev := h.Devices[id]
	if dev != nil {
		return fmt.Errorf("Device ID '%s' already exists", id)
	}

	maker, ok := h.makers[model]
	if !ok {
		return fmt.Errorf("Device Model '%s' not registered", model)
	}

	h.Devices[id] = &Device{model, name, false, maker(id, model, name)}
	h.storeDevices()

	return nil
}

func (h *Hub) apiCreate(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")
	err := h._create(id, model, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Hub) _delete(id string)  error {
	dev := h.Devices[id]
	if dev == nil {
		return fmt.Errorf("Device ID '%s' not found", id)
	}
	delete(h.Devices, id)
	h.storeDevices()

	var msg dean.Msg
	msg.Marshal(&dean.ThingMsgAbandon{Path: "abandon", Id: id})
	h.async <- &msg

	return nil
}

func (h *Hub) apiDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h._delete(id)
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

func (h *Hub) Run(i *dean.Injector) {
	h.restoreDevices()
	h.storeDevices()
	h.makeThingers()
	h.dumpDevices()

	for {
		select {
		case msg := <- h.async:
			i.Inject(msg)
		}
	}
}
