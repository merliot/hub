//go:build !tinygo

package gps

import (
	"bufio"
	"log"
	"fmt"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/gps/nmea"
	"github.com/tarm/serial"
)

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
