//go:build !tinygo

package move

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

func (m *Move) targetNew() {
	m.CompositeFs.AddFS(fs)
	m.templates = m.CompositeFs.ParseFS("template/*")
}

func (m *Move) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "state":
		common.ShowState(m.templates, w, m)
	default:
		m.Common.API(m.templates, w, r)
	}
}
