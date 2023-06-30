package sense

import (
	"embed"
	"net/http"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js index.html
var fs embed.FS

type Sense struct {
	*common.Common
	Lux int
	bh1750 *i2c.BH1750Driver
}

type Identity struct {
	Id    string
	Model string
	Name  string
}

type Update struct {
	Path string
	Lux int
}

func New(id, model, name string) dean.Thinger {
	println("NEW SENSE")
	return &Sense {
		Common: common.New(id, model, name).(*common.Common),
	}
}

func replyState(s *Sense) func(*dean.Msg) {
	s.Path = "state"
	return dean.ReplyState(s)
}

func (s *Sense) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     dean.SaveState(s),
		"get/state": replyState(s),
		"update":    dean.UpdateState(s),
	}
}

func (s *Sense) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ServeFS(fs, w, r)
}

func (s *Sense) Run(i *dean.Injector) {
	var msg dean.Msg
	var update = Update{Path: "update"}

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	s.bh1750 = i2c.NewBH1750Driver(adaptor)
	err := s.bh1750.Start()
	if err != nil {
		println(err.Error())
	}

	for {
		lux, err := s.bh1750.Lux()
		if err != nil {
			println(err.Error())
		}
		if lux != s.Lux {
			s.Lux = lux
			update.Lux  = lux
			i.Inject(msg.Marshal(update))
		}
		time.Sleep(time.Second)
	}

	select {}
}
