package sense

import (
	"embed"
	"net/http"
	"time"

	"github.com/merliot/dean"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js index.html
var fs embed.FS

type Sense struct {
	dean.Thing
	dean.ThingMsg
	Identity Identity
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
	return &Sense{
		Thing: dean.NewThing(id, model, name),
		Identity: Identity{id, model, name},
	}
}

func (s *Sense) saveState(msg *dean.Msg) {
	msg.Unmarshal(s)
}

func (s *Sense) getState(msg *dean.Msg) {
	s.Path = "state"
	msg.Marshal(s).Reply()
}

func (s *Sense) update(msg *dean.Msg) {
	msg.Unmarshal(s).Broadcast()
}

func (s *Sense) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     s.saveState,
		"get/state": s.getState,
		"update":    s.update,
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
