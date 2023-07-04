package led

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
)

//go:embed css js index.html
var fs embed.FS

type Led struct {
	*common.Common
}

func New(id, model, name string) dean.Thinger {
	println("NEW LED")
	return &Led{
		Common: common.New(id, model, name).(*common.Common),
	}
}

func reply(l *Led) func(*dean.Msg) {
	l.Path = "state"
	return dean.ThingReply(l)
}

func (l *Led) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     dean.ThingSave(l),
		"get/state": reply(l),
	}
}

func (l *Led) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.ServeFS(fs, w, r)
}
