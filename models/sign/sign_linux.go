//go:build !tinygo

package sign

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/merliot/dean"
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
	s.Common.API(s.templates, w, req)
}

func (s *Sign) run(i *dean.Injector) {
	select {}
}
