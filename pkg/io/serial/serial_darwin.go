//go:build darwin

package serial

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

func openPort(c *Config) (*Port, error) {
	f, err := os.OpenFile(c.Name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	fd := f.Fd()
	port := &Port{fd: int(fd), name: c.Name}

	// Get current serial port settings
	t := syscall.Termios{}
	r, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		port.Close()
		return nil, os.NewSyscallError("TIOCGETA", errno)
	}
	if r != 0 {
		port.Close()
		return nil, fmt.Errorf("tcgetattr failed")
	}

	// Set baud rate
	var speed uint64
	switch c.Baud {
	case 9600:
		speed = uint64(syscall.B9600)
	case 19200:
		speed = uint64(syscall.B19200)
	case 38400:
		speed = uint64(syscall.B38400)
	case 57600:
		speed = uint64(syscall.B57600)
	case 115200:
		speed = uint64(syscall.B115200)
	default:
		port.Close()
		return nil, fmt.Errorf("unsupported baud rate")
	}
	t.Ispeed = speed
	t.Ospeed = speed

	// Set data bits
	t.Cflag = (t.Cflag &^ syscall.CSIZE) | syscall.CS8

	// Set stop bits
	switch c.StopBits {
	case 1:
		t.Cflag &^= syscall.CSTOPB
	case 2:
		t.Cflag |= syscall.CSTOPB
	default:
		port.Close()
		return nil, fmt.Errorf("unsupported stop bits setting")
	}

	// Set parity
	switch c.Parity {
	case ParityNone:
		t.Cflag &^= syscall.PARENB
	case ParityOdd:
		t.Cflag |= syscall.PARENB
		t.Cflag |= syscall.PARODD
	case ParityEven:
		t.Cflag |= syscall.PARENB
		t.Cflag &^= syscall.PARODD
	default:
		port.Close()
		return nil, fmt.Errorf("unsupported parity setting")
	}

	// Enable receiver, ignore modem control lines
	t.Cflag |= syscall.CREAD | syscall.CLOCAL

	// Raw mode
	t.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ECHOE | syscall.ISIG
	t.Iflag &^= syscall.IXON | syscall.IXOFF | syscall.IXANY | syscall.INPCK | syscall.ISTRIP
	t.Oflag &^= syscall.OPOST

	// Set timeout
	if c.ReadTimeout > 0 {
		timeout := uint32(c.ReadTimeout / (time.Second / 10))
		t.Cc[syscall.VTIME] = uint8(timeout)
		t.Cc[syscall.VMIN] = 0
	} else {
		t.Cc[syscall.VMIN] = 1
		t.Cc[syscall.VTIME] = 0
	}

	// Apply settings
	r, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		port.Close()
		return nil, os.NewSyscallError("TIOCSETA", errno)
	}
	if r != 0 {
		port.Close()
		return nil, fmt.Errorf("tcsetattr failed")
	}

	return port, nil
}
