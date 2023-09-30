//go:build tinygo && nano_rp2040

package relays

import (
	"crypto/rand"
	"machine"

	"github.com/merliot/hub/models/common"
)

func (r *Relays) pins() common.GpioPins {
	return r.Targets["nano-rp2040"].GpioPins
}

// TODO: remove when RNG is working on rp2040

func init() {
	rand.Reader = &reader{}
}

type reader struct{}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	var randomByte uint32
	for i := range b {
		if i%4 == 0 {
			randomByte, err = machine.GetRNG()
			if err != nil {
				return n, err
			}
		} else {
			randomByte >>= 8
		}
		b[i] = byte(randomByte)
	}
	return len(b), nil
}
