package common

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/merliot/dean"
)

//go:embed css js template
var fs embed.FS

type Common struct {
	dean.Thing
	WebSocket   string `json:"-"`
	CompositeFs *dean.CompositeFS `json:"-"`
	templates   *template.Template
}

func New(id, model, name string) dean.Thinger {
	println("NEW COMMON")
	c := &Common{}
	c.Thing = dean.NewThing(id, model, name)
	c.CompositeFs = dean.NewCompositeFS()
	c.CompositeFs.AddFS(fs)
	c.templates = c.CompositeFs.ParseFS("template/*")
	return c
}

func RenderTemplate(templates *template.Template, w http.ResponseWriter, name string, data any) {
	tmpl := templates.Lookup(name)
	if tmpl != nil {
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Template '" + name + "' not found", http.StatusBadRequest)
	}
}

func (c *Common) API(templates *template.Template, w http.ResponseWriter, r *http.Request) {

	id, _, _ := c.Identity()
	c.WebSocket = scheme + r.Host + "/ws/" + id + "/"

	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "", "index.html":
		RenderTemplate(templates, w, "index.html", c)
	case "deploy.html":
		RenderTemplate(templates, w, "deploy.tmpl", c)
	case "deploy":
		c.deploy(templates, w, r)
	default:
		http.FileServer(http.FS(c.CompositeFs)).ServeHTTP(w, r)
	}
}
