//go:build !tinygo

package common

import (
	"bufio"
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/merliot/dean"
)

//go:embed *
var commonFs embed.FS

type commonOS struct {
	WebSocket   string            `json:"-"`
	PingPeriod  int               `json:"-"`
	CompositeFs *dean.CompositeFS `json:"-"`
	templates   *template.Template
}

func (c *Common) commonOSInit() {
	c.PingPeriod = 4
	//c.PingPeriod = 60
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
		http.Error(w, "Template '"+name+"' not found", http.StatusBadRequest)
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

func (c *Common) renderMarkdown(path string, w http.ResponseWriter) {
	file, err := c.CompositeFs.Open(path)
	if err != nil {
		http.Error(w, "File '"+path+"' not found", http.StatusNotFound)
		return
	}
	reader := bufio.NewReader(file)

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	md, _ := ioutil.ReadAll(reader)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	w.Header().Set("Content-Type", "text/html")
	w.Write(markdown.Render(doc, renderer))
}

// Set Content-Type: "text/plain" on go, css, and template files
var textFile = regexp.MustCompile("\\.(go|tmpl|js|css)$")

// Set Content-Type: "application/javascript" on js files
var scriptFile = regexp.MustCompile("\\.(go|tmpl|js|css)$")

// Markdown files get converted to html
var markdownFile = regexp.MustCompile("\\.md$")

func (c *Common) API(templates *template.Template, w http.ResponseWriter, r *http.Request) {

	id, _, _ := c.Identity()

	pingPeriod := strconv.Itoa(c.PingPeriod)
	c.WebSocket = wsScheme + r.Host + "/ws/" + id + "/?ping-period=" + pingPeriod

	path := r.URL.Path
	switch strings.TrimPrefix(path, "/") {
	case "", "index.html":
		RenderTemplate(templates, w, "index.tmpl", c)
	case "download":
		RenderTemplate(templates, w, "download.tmpl", c)
	case "info":
		RenderTemplate(templates, w, "info.tmpl", c)
	case "deploy":
		c.deploy(templates, w, r)
	case "code":
		c.showCode(templates, w, r)
	case "state":
		ShowState(templates, w, c)
	default:
		if markdownFile.MatchString(path) {
			c.renderMarkdown(path, w)
			return
		}
		if textFile.MatchString(path) {
			w.Header().Set("Content-Type", "text/plain")
		}
		if scriptFile.MatchString(path) {
			w.Header().Set("Content-Type", "application/javascript")
		}
		http.FileServer(http.FS(c.CompositeFs)).ServeHTTP(w, r)
	}
}

func (c *Common) Load() {
	bytes, err := os.ReadFile("devs/" + c.Id + ".json")
	if err == nil {
		json.Unmarshal(bytes, &c.DeployParams)
	}
}

func (c *Common) Save() {
	bytes, err := json.MarshalIndent(c.DeployParams, "", "\t")
	if _, err := os.Stat("devs/"); os.IsNotExist(err) {
		// If the directory doesn't exist, create it
		os.Mkdir("devs/", os.ModePerm)
	}
	if err == nil {
		os.WriteFile("devs/"+c.Id+".json", bytes, 0600)
	}
}
