//go:build !tinygo

package uv

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

func (u *Uv) targetNew() {
	u.CompositeFs.AddFS(fs)
	u.templates = u.CompositeFs.ParseFS("template/*")
}

func (u *Uv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "state":
		common.ShowState(u.templates, w, u)
	default:
		u.Common.API(u.templates, w, r)
	}
}
