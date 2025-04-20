//go:build windows

package serial

import (
	"fmt"
	"syscall"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateFile      = kernel32.NewProc("CreateFileW")
	procCloseHandle     = kernel32.NewProc("CloseHandle")
	procReadFile        = kernel32.NewProc("ReadFile")
	procWriteFile       = kernel32.NewProc("WriteFile")
	procGetCommState    = kernel32.NewProc("GetCommState")
	procSetCommState    = kernel32.NewProc("SetCommState")
	procSetCommTimeouts = kernel32.NewProc("SetCommTimeouts")
)

type _DCB struct {
	DCBlength  uint32
	BaudRate   uint32
	Flags      uint32
	wReserved  uint16
	XonLim     uint16
	XoffLim    uint16
	ByteSize   byte
	Parity     byte
	StopBits   byte
	XonChar    byte
	XoffChar   byte
	ErrorChar  byte
	EofChar    byte
	EvtChar    byte
	wReserved1 uint16
}

type _COMMTIMEOUTS struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

func openPort(c *Config) (*Port, error) {
	if c == nil {
		return nil, fmt.Errorf("nil config")
	}

	// Windows requires \\.\ prefix for COM ports above 9
	name := c.Name
	if len(name) > 4 && name[:4] == "COM" {
		if n, err := fmt.Sscanf(name[3:], "%d", new(int)); err == nil && n > 9 {
			name = `\\.\` + name
		}
	}

	h, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	handle, err := syscall.CreateFile(h,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0)
	if err != nil {
		return nil, err
	}

	return &Port{handle: handle, name: c.Name}, nil
}

func closePort(p *Port) error {
	handle, ok := p.handle.(syscall.Handle)
	if !ok {
		return fmt.Errorf("invalid handle type")
	}
	return syscall.CloseHandle(handle)
}

func readPort(p *Port, b []byte) (n int, err error) {
	handle, ok := p.handle.(syscall.Handle)
	if !ok {
		return 0, fmt.Errorf("invalid handle type")
	}

	var bytesRead uint32
	err = syscall.ReadFile(handle, b, &bytesRead, nil)
	return int(bytesRead), err
}

func writePort(p *Port, b []byte) (n int, err error) {
	handle, ok := p.handle.(syscall.Handle)
	if !ok {
		return 0, fmt.Errorf("invalid handle type")
	}

	var bytesWritten uint32
	err = syscall.WriteFile(handle, b, &bytesWritten, nil)
	return int(bytesWritten), err
}
