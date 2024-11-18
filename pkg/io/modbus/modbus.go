package modbus

import (
	"errors"
	"io"
	"time"
)

var (
	ErrTimeout = errors.New("Timeout")
)

type Modbus struct {
	// Read() should be blocking with timeout
	io.ReadWriter
}

func New(rw io.ReadWriter) Modbus {
	return Modbus{
		ReadWriter: rw,
	}
}

// calculateCRC calculates Modbus RTU CRC16
func calculateCRC(data []byte) []byte {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)

		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return []byte{byte(crc & 0xFF), byte(crc >> 8)}
}

func readRegisterReq(start, words uint16) []byte {
	req := []byte{1, 4,
		byte(start >> 8), byte(start & 0xff),
		byte(words >> 8), byte(words & 0xff)}
	crc := calculateCRC(req)
	req = append(req, crc...)
	return req
}

func (m *Modbus) ReadRegisters(start, words uint16) ([]byte, error) {
	time.Sleep(100 * time.Millisecond)

	req := readRegisterReq(start, words)
	_, err := m.Write(req)
	if err != nil {
		return nil, err
	}

	var want = int(5 + words*2)
	var res = make([]byte, want)
	var pos = 0

	for want > 0 {
		// Assuming Read() is blocking with timeout err
		n, err := m.Read(res[pos:])
		if err != nil {
			return nil, err
		}
		pos += n
		want -= n
	}

	return res[3 : pos-2], nil
}
