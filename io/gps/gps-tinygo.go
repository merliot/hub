//go:build tinygo

package gps

import (
	"machine"
)

type Gps struct {
	air530
}

func (g *Gps) Setup() error {
	return g.air530.Setup(machine.UART0, machine.UART0_TX_PIN,
		machine.UART0_RX_PIN, 9600)
}
