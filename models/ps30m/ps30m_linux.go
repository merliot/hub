//go:build !tinygo

package ps30m

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/merliot/hub/models/common"
)

//go:embed *
var fs embed.FS

type targetStruct struct {
	templates *template.Template
}

func (p *Ps30m) targetNew() {
	p.CompositeFs.AddFS(fs)
	p.templates = p.CompositeFs.ParseFS("template/*")
}

func (p *Ps30m) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "state":
		common.ShowState(p.templates, w, p)
	default:
		p.API(p.templates, w, r)
	}
}

func (p *Ps30m) write(buf []byte) (n int, err error) {
	// TODO
	return n, err
}

func (p *Ps30m) read(buf []byte) (n int, err error) {
	// TODO
	return n, err
}

