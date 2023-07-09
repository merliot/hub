package hub

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
)

//go:embed css html images js index.html
var fs embed.FS

type Device struct {
	Model string
	Name  string
	Online  bool
	thinger dean.Thinger
}

type Devices map[string]*Device        // keyed by id
type makers map[string]dean.ThingMaker // keyed by model

type Hub struct {
	*common.Common
	Devices
	makersMu sync.Mutex
	makers   makers
}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	return &Hub{
		Common: common.New(id, model, name).(*common.Common),
		Devices:  Devices{
			"88_ae_dd_0a_70_92": &Device{Model: "relays", Name: "Relays"},
			"d1":                &Device{Model: "relays", Name: "Relays01"},
			"d2":                &Device{Model: "relays", Name: "Relays02"},
		},
		makers:   makers{},
	}
}

func (h *Hub) Register(model string, maker dean.ThingMaker) {
	h.makersMu.Lock()
	defer h.makersMu.Unlock()
	h.makers[model] = maker
}

func (h *Hub) Unregister(model string) {
	h.makersMu.Lock()
	defer h.makersMu.Unlock()
	delete(h.makers, model)
}

func (h *Hub) getDevice(id, model, name string) (found *Device) {
	if dev, ok := h.Devices[id]; ok {
		if dev.Model == model && dev.Name == name {
			found = dev
		}
	}
	return
}

func (h *Hub) Make(id, model, name string) dean.Thinger {
	dev := h.getDevice(id, model, name)
	if dev == nil {
		return nil
	}
	return dev.thinger
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
	h.makersMu.Lock()
	defer h.makersMu.Unlock()
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

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.API(fs, w, r)
}

func (h *Hub) Run(i *dean.Injector) {
	h.storeDevices()
	h.restoreDevices()
	h.makeThingers()
	h.dumpDevices()
	select {}
}
