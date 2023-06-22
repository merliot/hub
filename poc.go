package poc

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/dean-lib/hub"
)

//go:embed css js index.html
var fs embed.FS

type Poc struct {
	*hub.Hub
}

func New(id, model, name string) dean.Thinger {
	println("NEW POC")
	return &Poc{
		Hub: hub.New(id, model, name).(*hub.Hub),
	}
}

func (p *Poc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ServeFS(fs, w, r)
}
