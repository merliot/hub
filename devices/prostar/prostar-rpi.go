//go:build rpi

package prostar

import (
	"time"

	"github.com/tarm/serial"
)

var defaultTty = "/dev/ttyUSB0"

type transport struct {
	*serial.Port
	tty string
}

func newTransport(tty string) *transport {
	if tty == "" {
		tty = defaultTty
	}
	return &transport{tty: tty}
}

func (t *transport) open() (err error) {
	if t.Port != nil {
		return
	}
	t.Port, err = serial.OpenPort(&serial.Config{
		Name:        t.tty,
		Baud:        9600,
		StopBits:    2,
		Parity:      serial.ParityNone,
		ReadTimeout: time.Second,
	})
	return
}

func (t *transport) Read(buf []byte) (int, error) {
	err := t.open()
	if err != nil {
		return 0, err
	}
	n, err := t.Port.Read(buf)
	if err != nil {
		t.Port.Close()
		t.Port = nil
		return 0, err
	}
	return n, nil
}

func (t *transport) Write(buf []byte) (int, error) {
	err := t.open()
	if err != nil {
		return 0, err
	}
	n, err := t.Port.Write(buf)
	if err != nil {
		t.Port.Close()
		t.Port = nil
		return 0, err
	}
	return n, nil
}
