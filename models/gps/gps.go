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

func (g *Gps) save(msg *dean.Msg) {
	msg.Unmarshal(g)
}

func (g *Gps) getState(msg *dean.Msg) {
	g.Path = "state"
	msg.Marshal(g).Reply()
}

func (g *Gps) update(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.save,
		"get/state": g.getState,
		"update":    g.update,
	}
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.API(fs, w, r)
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
