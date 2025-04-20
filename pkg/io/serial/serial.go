package serial

import (
	"io"
	"time"
)

type Config struct {
	Name        string
	Baud        int
	Size        byte // Byte size
	Parity      Parity
	StopBits    byte
	ReadTimeout time.Duration
}

type Port struct {
	handle interface{}
	name   string
}

type Parity byte

const (
	ParityNone  Parity = 0
	ParityOdd   Parity = 1
	ParityEven  Parity = 2
	ParityMark  Parity = 3
	ParitySpace Parity = 4
)

// OpenPort opens a serial port with the specified configuration
func OpenPort(c *Config) (*Port, error) {
	return openPort(c)
}

// Close closes the serial port
func (p *Port) Close() error {
	return closePort(p)
}

// Read reads from the serial port
func (p *Port) Read(b []byte) (n int, err error) {
	return readPort(p, b)
}

// Write writes to the serial port
func (p *Port) Write(b []byte) (n int, err error) {
	return writePort(p, b)
}

// Implement io.ReadWriteCloser interface
var _ io.ReadWriteCloser = &Port{}
