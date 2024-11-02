//go:build !tinygo

package gps

import (
	"bufio"
	"sync"

	"github.com/merliot/hub"
	"github.com/merliot/hub/io/gps/nmea"
	"github.com/tarm/serial"
)

type Gps struct {
	*serial.Port
	lat  float64
	long float64
	sync.RWMutex
}

func (g *Gps) Setup() (err error) {
	cfg := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}

	g.Lock()
	g.Port, err = serial.OpenPort(cfg)
	g.Unlock()

	if err != nil {
		return err
	}

	go g.scan()

	return nil
}

func (g *Gps) scan() {
	scanner := bufio.NewScanner(g.Port)
	for scanner.Scan() {
		//hub.LogDebug(scanner.Text())
		lat, long, err := nmea.ParseGLL(scanner.Text())
		if err != nil {
			//hub.LogError("Scan", "err", err)
			continue
		}
		g.Lock()
		g.lat, g.long = lat, long
		g.Unlock()
	}

	if err := scanner.Err(); err != nil {
		hub.LogError("Closing scan", "err", err)
	}

	g.Port.Close()

	g.Lock()
	g.Port = nil
	g.lat, g.long = 0.0, 0.0
	g.Unlock()
}

func (g *Gps) Location() (float64, float64, error) {
	g.RLock()
	if g.Port == nil {
		g.RUnlock()
		return 0.0, 0.0, g.Setup()
	}
	defer g.RUnlock()
	return g.lat, g.long, nil
}
