//go:build tinygo

package hp2430n

import (
	"fmt"
	"machine"
	"time"

	"github.com/merliot/dean"
)

type targetStruct struct {
	uart *machine.UART
}

func (h *Hp2430n) targetNew() {
	h.uart = machine.UART0
	h.uart.Configure(machine.UARTConfig{
		TX: machine.UART0_TX_PIN,
		RX: machine.UART0_RX_PIN,
		BaudRate: 9600,
	})
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

func readRegisterReq(start, count uint16) []byte {
	req := []byte{1, 3,
		byte(start >> 8), byte(start & 0xff),
		byte(count >> 8), byte(count & 0xff)}
	crc := calculateCRC(req)
	req = append(req, crc...)
	return req
}

func (h *Hp2430n) readRegisters(start, count uint16) ([]byte, error) {
	time.Sleep(100 * time.Millisecond)

	req := readRegisterReq(start, count)
	_, err := h.write(req)
	if err != nil {
		return nil, fmt.Errorf("Error writing request start %d count %d err %s",
			start, count, err)
	}

	var res = make([]byte, 5 + count*2)
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		n, err := h.read(res)
		if err != nil {
			return nil, fmt.Errorf("Error reading response start %d count %d err %s",
				start, count, err)
		}
		switch n {
		case 5:
			return nil, fmt.Errorf("Error reading response exception code %d", res[2])
		case int(5 + count*2):
			return res[3:n-2], nil
		}
	}

	return nil, fmt.Errorf("Timeout reading registers")
}

func (h *Hp2430n) readRegUint16(reg uint16) uint16 {
	res, err := h.readRegisters(reg, 1)
	if err != nil {
		println("Error reading register at address", reg, err.Error())
		return 0
	}
	return (uint16(res[0]) << 8) | uint16(res[1])
}

func (h *Hp2430n) readVoltage(reg uint16) float32 {
	voltage := h.readRegUint16(reg)
	return float32(voltage) * 0.1
}

func (h *Hp2430n) readCurrent(reg uint16) float32 {
	current := h.readRegUint16(reg)
	return float32(current) * 0.01
}

func (h *Hp2430n) readLoadInfo() uint16 {
	return h.readRegUint16(regLoadInfo)
}

func (h *Hp2430n) write(buf []byte) (n int, err error) {
	n, err = h.uart.Write(buf)
	return n, err
}

func (h *Hp2430n) read(buf []byte) (n int, err error) {
	return h.uart.Read(buf)
}

func (h *Hp2430n) Run(i *dean.Injector) {
	h.sample(i)
}
