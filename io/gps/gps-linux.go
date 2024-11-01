//go:build !tinygo

package gps

import (
	"bufio"

	"github.com/ietxaniz/delock"
	"github.com/merliot/hub"
	"github.com/merliot/hub/io/gps/nmea"
	"github.com/tarm/serial"
)

type Gps struct {
	*serial.Port
	lat  float64
	long float64
	delock.RWMutex
}

func (g *Gps) Setup() (err error) {
	cfg := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}

	lockId, err := g.Lock()
	if err != nil {
		panic(err)
	}
	g.Port, err = serial.OpenPort(cfg)
	g.Unlock(lockId)

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
		lockId, err := g.Lock()
		if err != nil {
			panic(err)
		}
		g.lat, g.long = lat, long
		g.Unlock(lockId)
	}

	if err := scanner.Err(); err != nil {
		hub.LogError("Closing scan", "err", err)
	}

	g.Port.Close()

	lockId, err := g.Lock()
	if err != nil {
		panic(err)
	}
	g.Port = nil
	g.lat, g.long = 0.0, 0.0
	g.Unlock(lockId)
}

func (g *Gps) Location() (float64, float64, error) {
	lockId, err := g.RLock()
	if err != nil {
		panic(err)
	}
	if g.Port == nil {
		g.RUnlock(lockId)
		return 0.0, 0.0, g.Setup()
	}
	defer g.RUnlock(lockId)
	return g.lat, g.long, nil
}
