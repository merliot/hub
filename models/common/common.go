package common

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/merliot/dean"
)

type Common struct {
	dean.Thing
	dean.ThingMsg
	Identity Identity
	WebSocket string `json:"-"`
}

type Identity struct {
	Id    string
	Model string
	Name  string
}

func New(id, model, name string) dean.Thinger {
	println("NEW COMMON")
	return &Common{
		Thing: dean.NewThing(id, model, name),
		Identity: Identity{id, model, name},
	}
}

func parseTemplate(data any, fs embed.FS, w http.ResponseWriter, file string) {
	tmpl, err := template.ParseFS(fs, file)
	if err != nil {
		println(err)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		println(err)
	}
}

func (c *Common) API(fs embed.FS, w http.ResponseWriter, r *http.Request) {
	c.WebSocket = scheme + r.Host + "/ws/" + c.Id() + "/"
	switch r.URL.Path {
	case "", "/":
		parseTemplate(c, fs, w, "index.html")
	default:
		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
	}
}
