package prime

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

//go:embed *
var fs embed.FS

type Device struct {
	Id     string
	Model  string
	Name   string
	Online bool
}

type Prime struct {
	*common.Common
	Device Device
	templates *template.Template
}

var targets = []string{"x86-64", "rpi"}

func New(id, model, name string) dean.Thinger {
	println("NEW PRIME")
	p := &Prime{}
	p.Common = common.New(id, model, name, targets).(*common.Common)
	p.CompositeFs.AddFS(fs)
	p.templates = p.CompositeFs.ParseFS("template/*")
	return p
}

func (p *Prime) getState(msg *dean.Msg) {
	p.Path = "state"
	msg.Marshal(p).Reply()
}

func (p *Prime) online(msg *dean.Msg, online bool) {
	p.Device.Online = online
	msg.Broadcast()
}

func (p *Prime) connect(online bool) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		p.online(msg, online)
	}
}

func (p *Prime) createdThing(msg *dean.Msg) {
	var create dean.ThingMsgCreated
	msg.Unmarshal(&create)
	p.Device = Device{Id: create.Id, Model: create.Model, Name: create.Name}
	create.Path = "created/device"
	msg.Marshal(&create).Broadcast()
}

func (p *Prime) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":     p.getState,
		"connected":     p.connect(true),
		"disconnected":  p.connect(false),
		"created/thing": p.createdThing,
	}
}

func (p *Prime) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.API(p.templates, w, r)
}
