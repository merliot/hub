//go:build !x86_64 && !tinygo && !rpi

package prostar

import (
	"math/rand"
)

type transport struct {
	start uint16
	words uint16
}

func newTransport(tty string) *transport {
	return &transport{}
}

func random(start, end float32) float32 {
	return round2(start + rand.Float32()*(end-start))
}

func (t *transport) Read(buf []byte) (n int, err error) {
	// simluate a Modbus request read on the device
	res := buf[3:]
	switch t.start {
	case regVerSw:
	case regAdcIa:
		copy(res[0:2], unf16(5.3))                // solar.Amps
		copy(res[2:4], unf16(random(14.1, 14.2))) // battery.Volts
		copy(res[4:6], unf16(random(15.3, 16.0))) // solar.Volts
		copy(res[6:8], unf16(random(12.6, 13.1))) // load.Volts
	case regAdcIl:
		copy(res[0:2], unf16(3.3))              // load.amps
		copy(res[2:4], unf16(0))                // battery.SenseVolts
		copy(res[4:6], unf16(0))                // battery.SlowVolts
		copy(res[6:8], unf16(random(1.2, 1.4))) // battery.SlowNetAmps
	case regChargeState:
		copy(res[0:2], unswap(8)) // charge state
	case regLoadState:
		copy(res[0:2], unswap(1)) // load state
		copy(res[2:4], unswap(0)) // load fault
	case regVbMinDaily:
		copy(res[0:2], unf16(13.94)) // min batt volts
		copy(res[2:4], unf16(14.30)) // max batt volts
		copy(res[4:6], unf16(15.82)) // daily system charge
		copy(res[6:8], unf16(17.30)) // daily load
	}
	return int(5 + t.words*2), nil
}

func (t *transport) Write(buf []byte) (n int, err error) {
	// get start and words from Modbus request
	t.start = (uint16(buf[2]) << 8) | uint16(buf[3])
	t.words = (uint16(buf[4]) << 8) | uint16(buf[5])
	return n, nil
}
