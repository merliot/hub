//go:build darwin
// +build darwin

package serial

import (
	"fmt"
	"time"
)

func openPort(name string, baud int, databits byte, parity Parity, stopbits StopBits, readTimeout time.Duration) (p *Port, err error) {
	// TODO
	return &Port{}, fmt.Errorf("not implemented")
}

type Port struct {
}

func (p *Port) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}

func (p *Port) Write(b []byte) (n int, err error) {
	return 0, fmt.Errorf("not implemented")
}

func (p *Port) Flush() error {
	return fmt.Errorf("not implemented")
}

func (p *Port) Close() (err error) {
	return fmt.Errorf("not implemented")
}
