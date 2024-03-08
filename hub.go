package hub

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/merliot/dean"
	"github.com/merliot/device"
)

//go:embed css go.mod images js template
var fs embed.FS

type Model struct {
	modeler          device.Modeler
	Icon             string
	DescHtml         template.HTML
	SupportedTargets string
}

type Models map[string]Model // keyed by model

type Child struct {
	Model          string
	Name           string
	DeployParams   string `json:"DeployParams,omitempty"`
	Online         bool   `json:"-"`
	device.Devicer `json:"-"`
}

type Children map[string]*Child // keyed by id

type Hub struct {
	*device.Device
	Version string
	Models  `json:"-"`
	Children
	server    *dean.Server
	templates *template.Template
}

var targets = []string{"x86-64", "rpi"}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	h := &Hub{}
	h.Device = device.New(id, model, name, fs, targets).(*device.Device)
	h.Version = version
	h.Models = make(Models)
	h.Children = make(Children)
	h.CompositeFs.AddFS(fs)
	h.templates = h.CompositeFs.ParseFS("template/*")
	return h
}

func (h *Hub) SetServer(server *dean.Server) {
	h.server = server
}

func (h *Hub) RegisterModel(model string, maker dean.ThingMaker) {
	if h.server != nil {
		h.server.RegisterModel(model, maker)
	}
	modeler := maker("proto", model, "proto").(device.Modeler)
	h.Models[model] = Model{
		modeler:          modeler,
		Icon:             base64.StdEncoding.EncodeToString(modeler.Icon()),
		DescHtml:         template.HTML(modeler.DescHtml()),
		SupportedTargets: modeler.SupportedTargets(),
	}
}

func (h *Hub) GenerateUf2s(dir string) error {
	for _, model := range h.Models {
		if err := model.modeler.GenerateUf2s(dir); err != nil {
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
	h.Children[child.Id] = &Child{Model: child.Model, Name: child.Name}
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

	h.Children[id].Devicer = device
}

func (h *Hub) LoadDevices(devices string) {
	var children Children
	json.Unmarshal([]byte(devices), &children)
	for id, child := range children {
		thinger, err := h.server.CreateThing(id, child.Model, child.Name)
		if err != nil {
			fmt.Printf("Skipping: error creating device Id '%s': %s\n", id, err)
			continue
		}
		h.loadDevice(thinger, id, child.DeployParams)
	}
}
