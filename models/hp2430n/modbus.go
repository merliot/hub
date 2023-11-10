package hp2430n

import (
	"fmt"
	"time"
)

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
	req := []byte{1, 3,
		byte(start >> 8), byte(start & 0xff),
		byte(words >> 8), byte(words & 0xff)}
	crc := calculateCRC(req)
	req = append(req, crc...)
	return req
}

func (h *Hp2430n) readRegisters(start, words uint16) ([]byte, error) {
	time.Sleep(100 * time.Millisecond)

	req := readRegisterReq(start, words)
	_, err := h.write(req)
	if err != nil {
		return nil, fmt.Errorf("Error writing request start %d words %d err %s",
			start, words, err)
	}

	var res = make([]byte, 5+words*2)
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		n, err := h.read(res)
		if err != nil {
			return nil, fmt.Errorf("Error reading response start %d words %d err %s",
				start, words, err)
		}
		switch n {
		case 5:
			return nil, fmt.Errorf("Error reading response exception code %d", res[2])
		case int(5 + words*2):
			// TODO validate CRC
			return res[3 : n-2], nil
		}
	}

	return nil, fmt.Errorf("Timeout reading registers")
}
