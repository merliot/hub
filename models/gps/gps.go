package gps

import (
	"bufio"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
	"github.com/merliot/hub/models/gps/nmea"
	"github.com/tarm/serial"
)

//go:embed css js template
var fs embed.FS

type Gps struct {
	*common.Common
	Lat  float64
	Long float64
	demo bool
	templates *template.Template
	ttyDevice string
	ttyBaud   int
}

type Update struct {
	Path string
	Lat  float64
	Long float64
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	g := &Gps{}
	g.Common = common.New(id, model, name).(*common.Common)
	g.CompositeFs.AddFS(fs)
	g.templates = g.CompositeFs.ParseFS("template/*")
	g.ttyDevice = "/dev/ttyUSB0"
	g.ttyBaud = 9600
	return g
}

func (g *Gps) SetSerial(dev string, baud int) {
	g.ttyDevice = dev
	g.ttyBaud = baud
}

func (g *Gps) save(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
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

func (g *Gps) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/deploy?target={target}\n"))
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "api":
		g.api(w, r)
	default:
		g.API(g.templates, w, r)
	}
}
func (g *Gps) Run(i *dean.Injector) {

	if g.demo {
		g.runDemo(i)
	}

	var msg dean.Msg
	var update = Update{Path: "update"}

	cfg := &serial.Config{Name: g.ttyDevice, Baud: g.ttyBaud}
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
		dist := int(distance(lat, long, g.Lat, g.Long) * 100.0) // cm
		if dist < 200 /*cm*/ {
			continue
		}
		g.Lat, g.Long = lat, long
		update.Lat, update.Long = lat, long
		i.Inject(msg.Marshal(update))
	}

	log.Fatal(fmt.Errorf("Disconnected from %s", g.ttyDevice))
}
