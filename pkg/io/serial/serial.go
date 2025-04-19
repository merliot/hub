package serial

import (
	"io"
	"syscall"
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
	fd   int
	name string
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
	return syscall.Close(p.fd)
}

// Read reads from the serial port
func (p *Port) Read(b []byte) (n int, err error) {
	return syscall.Read(p.fd, b)
}

// Write writes to the serial port
func (p *Port) Write(b []byte) (n int, err error) {
	return syscall.Write(p.fd, b)
}

// Implement io.ReadWriteCloser interface
var _ io.ReadWriteCloser = &Port{}
