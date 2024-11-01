//go:build tinygo

package gps

import (
	"machine"
	"time"

	"github.com/ietxaniz/delock"
	"github.com/merliot/hub/io/gps/nmea"
)

type air530 struct {
	uart *machine.UART
	lat  float64
	long float64
	delock.RWMutex
	buf [128]byte
}

func (a *air530) Setup(uart *machine.UART, tx, rx machine.Pin, baudrate uint32) error {
	a.uart = uart
	a.uart.Configure(machine.UARTConfig{TX: tx, RX: rx, BaudRate: baudrate})
	go a.scan()
	return nil
}

func (a *air530) process(raw string) {
	lat, long, err := nmea.ParseGLL(raw)
	if err == nil {
		lockId, err := a.Lock()
		if err != nil {
			panic(err)
		}
		defer a.Unlock(lockId)
		a.lat, a.long = lat, long
	}
}

func (a *air530) scan() {
	i := 0
	for {
		for a.uart.Buffered() > 0 {
			b, _ := a.uart.ReadByte()
			switch b {
			case '\r':
				a.process(string(a.buf[:i]))
				i = 0
			case '\n':
			default:
				a.buf[i] = b
				i++
				if i == len(a.buf) {
					i = 0
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (a air530) Location() (float64, float64, error) {
	lockId, err := a.RLock()
	if err != nil {
		panic(err)
	}
	defer a.RUnlock(lockId)
	return a.lat, a.long, nil
}
