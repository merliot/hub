package modela

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed css js images index.html
var fs embed.FS

type Modela struct {
	dean.Thing
	dean.ThingMsg
}

func New(id, model, name string) dean.Thinger {
	println("NEW MODELA")
	return &Modela{
		Thing: dean.NewThing(id, model, name),
	}
}

func (m *Modela) saveState(msg *dean.Msg) {
	msg.Unmarshal(m)
}

func (m *Modela) getState(msg *dean.Msg) {
	m.Path = "state"
	msg.Marshal(m).Reply()
}

func (m *Modela) update(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
}

func (m *Modela) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     m.saveState,
		"get/state": m.getState,
		"update":    m.update,
	}
}

func (m *Modela) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeFS(fs, w, r)
}
