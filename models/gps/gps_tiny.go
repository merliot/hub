//go:build tinygo

package gps

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	_ "github.com/merliot/dean/tinynet"
	"github.com/merliot/hub/models/gps/nmea"
)

type targetStruct struct {
}

func (g *Gps) targetNew() {
}

var (
	uart     = machine.UART0
	tx       = machine.UART0_TX_PIN
	rx       = machine.UART0_RX_PIN
	baudrate = 9600
)

func (g *Gps) parse(text string) (float64, float64, bool) {
	lat, long, err := nmea.ParseGLL(text)
	if err != nil {
		return 0, 0, false
	}
	return lat, long, true
}

func (g *Gps) moved(lat, long float64, radius int /* cm */) bool {
	dist := int(distance(lat, long, g.Lat, g.Long) * 100.0) // cm
	return dist >= radius
}

func (g *Gps) run(inj *dean.Injector) {

	var msg dean.Msg
	var update = Update{Path: "update"}
	var buf [128]byte

	uart.Configure(machine.UARTConfig{TX: tx, RX: rx, BaudRate: uint32(baudrate)})

	for i := 0; ; {
		for uart.Buffered() > 0 {
			b, _ := uart.ReadByte()
			switch b {
			case '\r':
				lat, long, good := g.parse(string(buf[:i]))
				if good && g.moved(lat, long, 200) {
					update.Lat, update.Long = lat, long
					inj.Inject(msg.Marshal(update))
				}
				i = 0
			case '\n':
			default:
				buf[i] = b
				i++
				if i == len(buf) {
					i = 0
				}
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
}
