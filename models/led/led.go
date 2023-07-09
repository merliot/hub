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

func (l *Led) save(msg *dean.Msg) {
	msg.Unmarshal(l)
}

func (l *Led) getState(msg *dean.Msg) {
	l.Path = "state"
	msg.Marshal(l).Reply()
}

func (l *Led) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     l.save,
		"get/state": l.getState,
	}
}

func (l *Led) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.API(fs, w, r)
}
