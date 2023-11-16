package prime

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

//go:embed *
var fs embed.FS

type Prime struct {
	*common.Common
	thinger dean.Thinger
	templates *template.Template
}

var targets = []string{"x86-64", "rpi"}

func New(thing dean.Thinger) dean.Thinger {
	println("NEW PRIME")
	p := &Prime{}
	p.Common = common.New("p1", "prime", "p1", targets).(*common.Common)
	p.thinger = thinger
	p.CompositeFs.AddFS(fs)
	p.templates = p.CompositeFs.ParseFS("template/*")
	return p
}

func (p *Prime) getState(msg *dean.Msg) {
	p.Path = "state"
	msg.Marshal(p).Reply()
}

func (p *Prime) online(msg *dean.Msg, online bool) {
	var thing dean.ThingMsgConnect
	msg.Unmarshal(&thing)
	fmt.Printf("%+v\r\n", thing)
	msg.Broadcast()
}

func (p *Prime) connect(online bool) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		p.online(msg, online)
	}
}

func (p *Prime) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":     p.getState,
		"connected":     p.connect(true),
		"disconnected":  p.connect(false),
	}
}

func (p *Prime) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.API(p.templates, w, r)
}
