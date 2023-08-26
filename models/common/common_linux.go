//go:build !tinygo

package common

import (
	"encoding/json"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"github.com/merliot/dean"
)

//go:embed *
var commonFs embed.FS

type commonOS struct {
	WebSocket   string `json:"-"`
	CompositeFs *dean.CompositeFS `json:"-"`
	templates   *template.Template
}

func (c *Common) commonOSInit() {
	c.CompositeFs = dean.NewCompositeFS()
	c.CompositeFs.AddFS(commonFs)
	c.templates = c.CompositeFs.ParseFS("template/*")
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

func (c *Common) showCode(templates *template.Template, w http.ResponseWriter, r *http.Request) {
	// Retrieve top-level entries
	entries, _ := fs.ReadDir(c.CompositeFs, ".")
	// Collect entry names
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	w.Header().Set("Content-Type", "text/html")
	RenderTemplate(templates, w, "code.tmpl", names)
}

func ShowState(templates *template.Template, w http.ResponseWriter, data any) {
	state, _ := json.MarshalIndent(data, "", "\t")
	RenderTemplate(templates, w, "state.tmpl", string(state))
}

// Set Content-Type: "text/plain" on go, js, css, and template files
var textFile = regexp.MustCompile("\\.(go|tmpl|js|css)$")

func (c *Common) API(templates *template.Template, w http.ResponseWriter, r *http.Request) {

	id, _, _ := c.Identity()
	c.WebSocket = scheme + r.Host + "/ws/" + id + "/"

	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "", "index.html":
		RenderTemplate(templates, w, "index.tmpl", c)
	case "deploy-dialog":
		RenderTemplate(templates, w, "deploy.tmpl", c)
	case "deploy":
		c.deploy(templates, w, r)
	case "code":
		c.showCode(templates, w, r)
	case "state":
		ShowState(templates, w, c)
	default:
		if textFile.MatchString(r.RequestURI) {
			w.Header().Set("Content-Type", "text/plain")
		}
		http.FileServer(http.FS(c.CompositeFs)).ServeHTTP(w, r)
	}
}
