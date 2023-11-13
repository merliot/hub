//go:build !tinygo

package hp2430n

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

func (h *Hp2430n) targetNew() {
	h.CompositeFs.AddFS(fs)
	h.templates = h.CompositeFs.ParseFS("template/*")
}

func (h *Hp2430n) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "state":
		common.ShowState(h.templates, w, h)
	default:
		h.API(h.templates, w, r)
	}
}

func (h *Hp2430n) write(buf []byte) (n int, err error) {
	// TODO
	return n, err
}

func (h *Hp2430n) read(buf []byte) (n int, err error) {
	// TODO
	return n, err
}
