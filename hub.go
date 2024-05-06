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
	"github.com/merliot/device/models"
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
	Version       string
	Demo          bool
	models.Models `json:"-"`
	Children
	server *dean.Server
}

var targets = []string{"x86-64", "rpi"}

func New(id, model, name string) dean.Thinger {
	fmt.Println("NEW HUB\r")
	h := &Hub{}
	h.Device = device.New(id, model, name, fs, targets).(*device.Device)
	h.Version = version
	h.Models = make(models.Models)
	h.Children = make(Children)
	return h
}

func (h *Hub) SetServer(server *dean.Server) {
	h.server = server
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

func (h *Hub) RegisterModel(model string, maker dean.ThingMaker) {
	proto := models.New(model, maker)
	h.Models[model] = proto
	if h.server != nil {
		h.server.RegisterModel(model, maker)
	}
}

func (h *Hub) GenerateUf2s(dir string) error {
	for _, model := range h.Models {
		if err := model.Modeler.GenerateUf2s(dir); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hub) getState(msg *dean.Msg) {
	h.Path = "state"
	msg.Marshal(h).Reply()
}

func (h *Hub) online(msg *dean.Msg, online bool) {
	var thing dean.ThingMsgConnect
	msg.Unmarshal(&thing)

	if child, ok := h.Children[thing.Id]; ok {
		child.Online = online
		msg.Broadcast()
	}
}

func (h *Hub) connect(online bool) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		h.online(msg, online)
	}
}

func (h *Hub) createdThing(msg *dean.Msg) {
	var child dean.ThingMsgCreated
	msg.Unmarshal(&child)
	h.Children[child.Id] = &Child{Id: child.Id, Model: child.Model, Name: child.Name}
	child.Path = "created/device"
	msg.Marshal(&child).Broadcast()
}

func (h *Hub) deletedThing(msg *dean.Msg) {
	var child dean.ThingMsgDeleted
	msg.Unmarshal(&child)
	delete(h.Children, child.Id)
	child.Path = "deleted/device"
	msg.Marshal(&child).Broadcast()
}

func (h *Hub) restart(msg *dean.Msg) {
	fmt.Println("RESTART")
	os.Exit(0)
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":     h.getState,
		"connected":     h.connect(true),
		"disconnected":  h.connect(false),
		"created/thing": h.createdThing,
		"deleted/thing": h.deletedThing,
		"restart":       h.restart,
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

func (h *Hub) LoadDevices(devices string) {
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
