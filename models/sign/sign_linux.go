//go:build !tinygo

package sign

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

//go:embed *
var fs embed.FS

type targetStruct struct {
	templates *template.Template
}

func (s *Sign) targetNew() {
	s.CompositeFs.AddFS(fs)
	s.templates = s.CompositeFs.ParseFS("template/*")
}

func (s *Sign) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch strings.TrimPrefix(req.URL.Path, "/") {
	case "state":
		common.ShowState(s.templates, w, s)
	default:
		s.Common.API(s.templates, w, req)
	}
}

func (s *Sign) refresh() {
}

func (s *Sign) store() {
}

func (s *Sign) run(i *dean.Injector) {
	select {}
}
