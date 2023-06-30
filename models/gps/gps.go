package gps

import (
	"bufio"
	"embed"
	"log"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/dean-lib/gps"
	"github.com/merliot/dean-lib/gps/nmea"
	"github.com/merliot/sw-poc/models/common"
	"github.com/tarm/serial"
)

//go:embed css js index.html
var fs embed.FS

type Gps struct {
	*common.Common
	Lat  float64
	Long float64
}

type Identity struct {
	Id    string
	Model string
	Name  string
}

type Update struct {
	Path string
	Lat  float64
	Long float64
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	return &Gps{
		Common: common.New(id, model, name).(*common.Common),
	}
}

func replyState(g *Gps) func(*dean.Msg) {
	g.Path = "state"
	return dean.ReplyState(g)
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     dean.SaveState(g),
		"get/state": replyState(g),
		"update":    dean.UpdateState(g),
	}
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.ServeFS(fs, w, r)
}

func (g *Gps) Run(i *dean.Injector) {
	var msg dean.Msg
	var update = Update{Path: "update"}

	cfg := &serial.Config{Name: "/dev/ttyS0", Baud: 9600}
	ser, err := serial.OpenPort(cfg)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(ser)
	for scanner.Scan() {
		lat, long, err := nmea.ParseGLL(scanner.Text())
		if err != nil {
			continue
		}
		dist := int(gps.Distance(lat, long, g.Lat, g.Long) * 100.0) // cm
		if dist < 200 /*cm*/ {
			continue
		}
		g.Lat, g.Long = lat, long
		update.Lat, update.Long  = lat, long
		i.Inject(msg.Marshal(update))
	}
}
