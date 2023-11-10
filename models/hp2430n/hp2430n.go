package hp2430n

import (
	"fmt"
	"strings"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

const (
	regMaxVoltage      = 0x000A
	regBatteryCapacity = 0x0100
	regLoadInfo        = 0x0120
)

type System struct {
	MaxVolts      uint8
	ChargeAmps    uint8
	DischargeAmps uint8
	ProductType   uint8
	Model         string
	SWVersion     string
	HWVersion     string
	Serial        string
	Temp          uint8 // deg C
}

type Battery struct {
	SOC         uint8
	Volts       float32
	Amps        float32
	Temp        uint8 // deg C
	ChargeState string
}

type LoadInfo struct {
	Volts      float32
	Amps       float32
	Status     bool
	Brightness uint8
}

type Solar struct {
	Volts float32
	Amps  float32
}

type Hp2430n struct {
	*common.Common
	System   System
	Battery  Battery
	LoadInfo LoadInfo
	Solar    Solar
	targetStruct
}

var targets = []string{"x86-64", "rpi", "nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW HP2430N")
	h := &Hp2430n{}
	h.Common = common.New(id, model, name, targets).(*common.Common)
	h.targetNew()
	return h
}

func (h *Hp2430n) save(msg *dean.Msg) {
	msg.Unmarshal(h).Broadcast()
}

func (h *Hp2430n) getState(msg *dean.Msg) {
	h.Path = "state"
	msg.Marshal(h).Reply()
}

/*
func (h *Hp2430n) updateStatus(msg *dean.Msg) {
	msg.Unmarshal(h).Broadcast()
}
*/

func (h *Hp2430n) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     h.save,
		"get/state": h.getState,
	}
}

func version(b []byte) string {
	return fmt.Sprintf("%02d.%02d.%02d", b[1], b[2], b[3])
}

func serial(b []byte) string {
	return fmt.Sprintf("%02X%02X-%02X%02X", b[0], b[1], b[2], b[3])
}

func volts(b []byte) float32 {
	raw := (uint16(b[0]) << 8) | uint16(b[1])
	return float32(raw) * 0.1
}

func amps(b []byte) float32 {
	raw := (uint16(b[0]) << 8) | uint16(b[1])
	return float32(raw) * 0.01
}

func chargeState(b byte) string {
	switch b {
	case 0:
		return "Deactivated"
	case 1:
		return "Activated"
	case 2:
		return "Mode MPPT"
	case 3:
		return "Mode Equalizing"
	case 4:
		return "Mode Boost"
	case 5:
		return "Mode Float"
	case 6:
		return "Current Limiting (Overpower)"
	}
	return "Unknown"
}

func (h *Hp2430n) readSystem(s *System) error {
	// System Info (34 bytes)
	regs, err := h.readRegisters(regMaxVoltage, 17)
	if err != nil {
		return err
	}
	s.MaxVolts = uint8(regs[0])
	s.ChargeAmps = uint8(regs[1])
	s.DischargeAmps = uint8(regs[2])
	s.ProductType = uint8(regs[3])
	s.Model = strings.ReplaceAll(string(regs[4:20]), "\000", "")
	s.SWVersion = version(regs[20:24])
	s.HWVersion = version(regs[24:28])
	s.Serial = serial(regs[28:32])
	// skip dev addr regs[32:34]
	return nil
}

func (h *Hp2430n) readDynamic(c *System, b *Battery, l *LoadInfo, s *Solar) error {

	// Controller Dynamic Info (20 bytes)
	regs, err := h.readRegisters(regBatteryCapacity, 10)
	if err != nil {
		return err
	}
	// reserved regs[0]
	b.SOC = uint8(regs[1])
	b.Volts = volts(regs[2:4])
	b.Amps = amps(regs[4:6])
	c.Temp = uint8(regs[6])
	b.Temp = uint8(regs[7])
	l.Volts = volts(regs[8:10])
	l.Amps = amps(regs[10:12])
	// skip load power regs[12:14]
	s.Volts = volts(regs[14:16])
	s.Amps = amps(regs[16:18])
	// skip solar power regs[18:20]

	// Load Information (2 bytes)
	regs, err = h.readRegisters(regLoadInfo, 1)
	if err != nil {
		return err
	}
	l.Status = (regs[0] & 0x80) == 0x80
	l.Brightness = uint8(regs[0] & 0x7F)
	b.ChargeState = chargeState(regs[1])

	return nil
}

func (h *Hp2430n) Run(i *dean.Injector) {

	h.Lock()
	if err := h.readSystem(&h.System); err != nil {
		println(err.Error())
	}
	if err := h.readDynamic(&h.System, &h.Battery, &h.LoadInfo, &h.Solar); err != nil {
		println(err.Error())
	}
	h.Unlock()

	select {}

	/*
		h.readDynamic()
		h.readDaily()
		h.readHistorical()

		for {
			regs, err := h.readRegisters(regBatteryVoltage, 8)
			println("len(regs)", len(regs))
			if err != nil {
				println("Error reading registers", err.Error())
				continue
			}
			info, err := h.readRegisters(regLoadInfo, 1)
			println("len(info)", len(info))
			if err != nil {
				println("Error reading registers", err.Error())
				continue
			}
			regs = append(regs, info[0])

			if !slicesAreEqual(regs, h.lastRegs) {
			}

			time.Sleep(time.Second)
		}
	*/
}
