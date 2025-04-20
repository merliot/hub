//go:build darwin

package serial

import (
	"fmt"
	"syscall"
)

func openPort(c *Config) (*Port, error) {
	if c == nil {
		return nil, fmt.Errorf("nil config")
	}

	fd, err := syscall.Open(c.Name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	return &Port{handle: fd, name: c.Name}, nil
}

func closePort(p *Port) error {
	fd, ok := p.handle.(int)
	if !ok {
		return fmt.Errorf("invalid handle type")
	}
	return syscall.Close(fd)
}

func readPort(p *Port, b []byte) (n int, err error) {
	fd, ok := p.handle.(int)
	if !ok {
		return 0, fmt.Errorf("invalid handle type")
	}
	return syscall.Read(fd, b)
}

func writePort(p *Port, b []byte) (n int, err error) {
	fd, ok := p.handle.(int)
	if !ok {
		return 0, fmt.Errorf("invalid handle type")
	}
	return syscall.Write(fd, b)
}
