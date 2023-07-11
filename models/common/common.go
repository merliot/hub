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

func parseTemplate(data any, tmpls *template.Template, w http.ResponseWriter, name string) {
	err := tmpls.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (c *Common) API(fs embed.FS, tmpls *template.Template, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "", "/":
		c.WebSocket = scheme + r.Host + "/ws/" + c.Id() + "/"
		parseTemplate(c, tmpls, w, "index.html")
	default:
		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
	}
}
