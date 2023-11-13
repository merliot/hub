//go:build tinygo

package ps30m

import (
	"machine"
)

type targetStruct struct {
	uart *machine.UART
}

func (p *Ps30m) targetNew() {
	p.uart = machine.UART0
	p.uart.Configure(machine.UARTConfig{
		TX:       machine.UART0_TX_PIN,
		RX:       machine.UART0_RX_PIN,
		BaudRate: 9600,
	})
	p.uart.SetFormat(8, 2, machine.ParityNone) // 8N2
}

func (p *Ps30m) write(buf []byte) (n int, err error) {
	n, err = p.uart.Write(buf)
	return n, err
}

func (p *Ps30m) read(buf []byte) (n int, err error) {
	return p.uart.Read(buf)
}
