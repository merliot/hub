//go:build tinygo

package prostar

import (
	"errors"
	"machine"
	"time"

	"github.com/merliot/hub/pkg/device"
)

func (p *prostar) GetConfig() device.Config {
	return device.Config{
		Model:      "prostar",
		State:      p,
		PollPeriod: pollPeriod,
	}
}

type transport struct {
	*machine.UART
}

func newTransport(tty string) *transport {
	t := transport{machine.DefaultUART}
	t.UART.Configure(machine.UARTConfig{
		BaudRate: 9600,
		TX:       machine.UART_TX_PIN,
		RX:       machine.UART_RX_PIN,
	})
	t.UART.SetFormat(8, 2, machine.ParityNone) // 8N2
	return &t
}

func (t *transport) Read(buf []byte) (int, error) {

	// The tinygo UART is non-blocking.
	// Make our Read blocking with 1 sec timeout.

	timeout := time.Now().Add(time.Second)

	for time.Now().Before(timeout) {
		n, err := t.UART.Read(buf)
		if err != nil {
			return 0, err
		}
		if n > 0 {
			return n, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return 0, errors.New("EOF")
}
