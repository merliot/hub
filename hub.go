package hub

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/device"
)

//go:embed css images js template
var fs embed.FS

const (
	dirChildren  = "children/"
	fileChildren = "children.json"
)

type Model struct {
	Icon             string
	DescHtml         template.HTML
	SupportedTargets string
}

type Models map[string]Model // keyed by model

type Child struct {
	Model  string
	Name   string
	Online bool `json:"-"`
}

type Children map[string]*Child // keyed by id

type Hub struct {
	*device.Device
	Models `json:"-"`
	Children
	server    *dean.Server
	gitKey    string
	gitRemote string
	gitAuthor string
	templates *template.Template
}

var targets = []string{"x86-64", "rpi"}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	h := &Hub{}
	h.Device = device.New(id, model, name, targets).(*device.Device)
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
	if h.server == nil {
		return
	}
	h.server.RegisterModel(model, maker)
	modeler := maker("proto", model, "proto").(device.Modeler)
	h.Models[model] = Model{
		Icon:             base64.StdEncoding.EncodeToString(modeler.Icon()),
		DescHtml:         template.HTML(modeler.DescHtml()),
		SupportedTargets: modeler.SupportedTargets(),
	}
}

func (h *Hub) SetGit(remote, key, author string) {
	h.gitRemote = remote
	h.gitKey = key
	h.gitAuthor = author
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
	}

	msg.Broadcast()
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
	h.storeChildren()
}

func filePath(id string) string {
	return dirChildren + id + ".json"
}

func (h *Hub) deletedThing(msg *dean.Msg) {
	var child dean.ThingMsgDeleted
	msg.Unmarshal(&child)
	delete(h.Children, child.Id)
	child.Path = "deleted/device"
	msg.Marshal(&child).Broadcast()
	os.Remove(filePath(child.Id))
	h.storeChildren()
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

func (h *Hub) restoreChildren() {
	var children Children
	bytes, _ := os.ReadFile(fileChildren)
	json.Unmarshal(bytes, &children)
	for id, child := range children {
		thinger, err := h.server.CreateThing(id, child.Model, child.Name)
		if err != nil {
			fmt.Printf("Skipping: error creating device Id '%s': %s\n", id, err)
			continue
		}
		device := thinger.(device.Devicer)
		device.CopyWifiAuth(h.WifiAuth)
		device.Load(filePath(id))
	}
}

func (h *Hub) storeChildren() {
	bytes, _ := json.MarshalIndent(h.Children, "", "\t")
	os.WriteFile(fileChildren, bytes, 0600)
}

func (h *Hub) Run(i *dean.Injector) {
	h.restoreChildren()
	for {
		err := h.saveChildren()
		if err != nil {
			fmt.Println("saving children error:", err.Error)
		}
		time.Sleep(5 * time.Second)
	}
}
