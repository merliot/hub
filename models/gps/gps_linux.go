//go:build !tinygo

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

//go:embed *
var fs embed.FS

type gpsOS struct {
	templates *template.Template
	ttyDevice string
	ttyBaud   int
}

func (g *Gps) gpsOSNew() {
	g.CompositeFs.AddFS(fs)
	g.templates = g.CompositeFs.ParseFS("template/*")
	g.ttyDevice = "/dev/ttyUSB0"
	g.ttyBaud = 9600
}

func (g *Gps) SetSerial(dev string, baud int) {
	g.ttyDevice = dev
	g.ttyBaud = baud
}

func (g *Gps) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/deploy?target={target}\n"))
	w.Write([]byte("/state\n"))
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "api":
		g.api(w, r)
	case "state":
		common.ShowState(g.templates, w, g)
	default:
		g.Common.API(g.templates, w, r)
	}
}

func (g *Gps) run(i *dean.Injector) {

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
