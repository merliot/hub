package hub

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"

	"github.com/merliot/dean"
	"github.com/merliot/device"
)

//go:embed css go.mod images js template
var fs embed.FS

type Child struct {
	Id             string
	Model          string
	Name           string
	DeployParams   string `json:"DeployParams,omitempty"`
	Online         bool   `json:"-"`
	device.Devicer `json:"-"`
}

func (c *Child) ModelTitle() template.JS {
	return template.JS(strings.Title(c.Model))
}

type Children map[string]*Child // keyed by id

type Hub struct {
	*device.Device
	server  *dean.Server
	Version string
	Demo    bool
	Children
}

var targets = []string{"x86-64", "rpi"}

func New(id, model, name string) dean.Thinger {
	fmt.Println("NEW HUB\r")
	return &Hub{
		Device:   device.New(id, model, name, fs, targets).(*device.Device),
		Version:  version,
		Children: make(Children),
	}
}

func NewHub(id, model, name, user, passwd, port, devices string) dean.Thinger {
	hub := New(id, model, name).(*Hub)
	hub.server = dean.NewServer(hub, user, passwd, port)
	for model, maker := range models {
		hub.server.RegisterModel(model, maker)
	}
	hub.loadDevices(devices)
	return hub
}

func (h *Hub) Serve() {
	h.server.Run()
}

func (h *Hub) SetBackup(backup string) {
	if backup == "" {
		return
	}
	u, err := url.Parse(backup)
	if err != nil {
		fmt.Println(backup, "is not a valid URL:", err)
		return
	}
	ws := "ws://"
	if u.Scheme == "https" {
		ws = "wss://"
	}
	dialURL := ws + u.Host + "/ws/?ping-period=4"
	h.SetDialURLs(dialURL)
}

func (h *Hub) SetDemo(demo bool) {
	h.Locked = (demo == true)
	h.Demo = demo
}

func (h *Hub) GenerateUf2s(dir string) error {
	for _, model := range h.server.Models() {
		if modeler, ok := model.(device.Modeler); ok {
			if err := modeler.GenerateUf2s(dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Hub) getState(pkt *dean.Packet) {
	pkt.Path = "state"
	pkt.Marshal(h).Reply()
}

func (h *Hub) online(pkt *dean.Packet, online bool) {
	var thing dean.ThingMsgConnect
	pkt.Unmarshal(&thing)

	if child, ok := h.Children[thing.Id]; ok {
		child.Online = online
		pkt.Broadcast()
	}
}

func (h *Hub) connect(online bool) func(*dean.Packet) {
	return func(pkt *dean.Packet) {
		h.online(pkt, online)
	}
}

func (h *Hub) createdThing(pkt *dean.Packet) {
	var child dean.ThingMsgCreated
	pkt.Unmarshal(&child)
	h.Children[child.Id] = &Child{Id: child.Id, Model: child.Model, Name: child.Name}
	pkt.SetPath("created/device").Broadcast()
}

func (h *Hub) deletedThing(pkt *dean.Packet) {
	var child dean.ThingMsgDeleted
	pkt.Unmarshal(&child)
	delete(h.Children, child.Id)
	pkt.SetPath("deleted/device").Broadcast()
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

func (h *Hub) loadDevice(thinger dean.Thinger, id, deployParams string) {
	device := thinger.(device.Devicer)

	device.SetDeployParams(deployParams)
	device.CopyWifiAuth(h.WifiAuth)
	device.SetWsScheme(h.WsScheme)
	device.SetDialURLs(h.DialURLs)
	device.SetLocked(h.Locked)

	h.Children[id].Devicer = device

	if h.Demo {
		thinger.Setup()
		thinger.SetFlag(dean.ThingFlagMetal)
		thinger.SetOnline(true)
		injector := h.server.NewInjector(id)
		go thinger.Run(injector)
	}
}

func (h *Hub) loadDevices(devices string) {
	var children Children

	// If devices is empty, try loading from devices.json file
	if devices == "" {
		data, err := os.ReadFile("devices.json")
		if err != nil {
			return
		}
		devices = string(data)
	}

	err := json.Unmarshal([]byte(devices), &children)
	if err != nil {
		fmt.Printf("Error parsing devices: %s\n", err)
		return
	}

	for id, child := range children {
		thinger, err := h.server.CreateThing(id, child.Model, child.Name)
		if err != nil {
			fmt.Printf("Skipping: error creating device Id '%s': %s\n", id, err)
			continue
		}
		h.loadDevice(thinger, id, child.DeployParams)
	}
}
