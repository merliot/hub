package ps30m

import (
	"strconv"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
	"github.com/x448/float16"
)

const (
	reg_ver_sw          = 0x0000
	reg_adc_ic_f_shadow = 0x0010
)

type System struct {
	SWVersion     string
	BattVoltMulti uint16
}

type Controller struct {
	Amps float32
}

type Battery struct {
	Volts      float32
	Amps       float32
	SenseVolts float32
	SlowVolts  float32
	SlowAmps   float32
}

type LoadInfo struct {
	Volts float32
	Amps  float32
}

type Solar struct {
	Volts float32
	Amps  float32
}

type msgStatus struct {
	Path   string
	Status string
}

type msgSystem struct {
	Path   string
	System System
}

type msgController struct {
	Path       string
	Controller Controller
}

type msgBattery struct {
	Path    string
	Battery Battery
}

type msgLoadInfo struct {
	Path     string
	LoadInfo LoadInfo
}

type msgSolar struct {
	Path  string
	Solar Solar
}

type Ps30m struct {
	*common.Common
	Status string
	System     System
	Controller Controller
	Battery    Battery
	LoadInfo   LoadInfo
	Solar      Solar
	targetStruct
}

var targets = []string{"x86-64", "rpi", "nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW PS30M")
	p := &Ps30m{}
	p.Common = common.New(id, model, name, targets).(*common.Common)
	p.Status = "OK"
	p.targetNew()
	return p
}

func (p *Ps30m) save(msg *dean.Msg) {
	msg.Unmarshal(p).Broadcast()
}

func (p *Ps30m) getState(msg *dean.Msg) {
	p.Path = "state"
	msg.Marshal(p).Reply()
}

func (p *Ps30m) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":             p.save,
		"get/state":         p.getState,
		"update/status":     p.save,
		"update/system":     p.save,
		"update/controller": p.save,
		"update/battery":    p.save,
		"update/load":       p.save,
		"update/solar":      p.save,
	}
}

func swap(b []byte) uint16 {
	return (uint16(b[0]) << 8) | uint16(b[1])
}

func noswap(b []byte) uint16 {
	return (uint16(b[1]) << 8) | uint16(b[0])
}

func f16(b []byte) float32 {
	return float16.Float16(swap(b)).Float32()
}

// bcdToDecimal converts a BCD-encoded uint16 value to decimal.
func bcdToDecimal(bcd uint16) string {
	decimal := uint16(0)
	multiplier := uint16(1)

	for bcd > 0 {
		// Extract the rightmost 4 bits (a decimal digit) from BCD
		digit := bcd & 0xF

		// Add the decimal value of the digit to the result
		decimal += digit * multiplier

		// Move to the next decimal place
		multiplier *= 10

		// Shift BCD to the right by 4 bits
		bcd >>= 4
	}

	// Convert the decimal value to a string
	return strconv.FormatUint(uint64(decimal), 10)
}

func (p *Ps30m) readSystem(s *System) error {
	regs, err := p.readRegisters(reg_ver_sw, 2)
	if err != nil {
		return err
	}
	s.SWVersion = bcdToDecimal(swap(regs[0:2]))
	s.BattVoltMulti = noswap(regs[2:4])
	return nil
}

func (p *Ps30m) readDynamic(c *Controller, b *Battery, l *LoadInfo, s *Solar) error {

	regs, err := p.readRegisters(reg_adc_ic_f_shadow, 10)
	if err != nil {
		return err
	}

	// Filtered ADC
	c.Amps = f16(regs[0:2])
	s.Amps = f16(regs[2:4])
	b.Volts = f16(regs[4:6])
	s.Volts = f16(regs[6:8])
	l.Volts = f16(regs[8:10])
	b.Amps = f16(regs[10:12])
	l.Amps = f16(regs[12:14])
	b.SenseVolts = f16(regs[14:16])
	b.SlowVolts = f16(regs[16:18])
	b.SlowAmps = f16(regs[18:20])

	return nil
}

func (p *Ps30m) sendStatus(i *dean.Injector, newStatus string) {
	if p.Status == newStatus {
		return
	}

	var status = msgStatus{Path: "update/status"}
	var msg dean.Msg

	status.Status = newStatus
	i.Inject(msg.Marshal(status))
}

func (p *Ps30m) sendSystem(i *dean.Injector) {
	var system = msgSystem{Path: "update/system"}
	var msg dean.Msg

	// sendSystem blocks until we get a good system info read

	for {
		if err := p.readSystem(&system.System); err != nil {
			p.sendStatus(i, err.Error())
			continue
		}
		i.Inject(msg.Marshal(system))
		break
	}

	p.sendStatus(i, "OK")
}

func (p *Ps30m) sendDynamic(i *dean.Injector) {
	var controller = msgController{Path: "update/controller"}
	var battery = msgBattery{Path: "update/battery"}
	var loadInfo = msgLoadInfo{Path: "update/load"}
	var solar = msgSolar{Path: "update/solar"}
	var msg dean.Msg

	err := p.readDynamic(&controller.Controller, &battery.Battery,
		&loadInfo.LoadInfo, &solar.Solar)
	if err != nil {
		p.sendStatus(i, err.Error())
		return
	}

	// If anything has changed, send update msg(s)

	if controller.Controller != p.Controller {
		i.Inject(msg.Marshal(controller))
	}
	if battery.Battery != p.Battery {
		i.Inject(msg.Marshal(battery))
	}
	if loadInfo.LoadInfo != p.LoadInfo {
		i.Inject(msg.Marshal(loadInfo))
	}
	if solar.Solar != p.Solar {
		i.Inject(msg.Marshal(solar))
	}

	p.sendStatus(i, "OK")
}

func (p *Ps30m) Run(i *dean.Injector) {

	p.sendSystem(i)
	p.sendDynamic(i)
	//p.sendHourly(i)

	nextHour := time.Now().Add(time.Hour)
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		p.sendDynamic(i)
		if time.Now().After(nextHour) {
			//p.sendHourly(i)
			nextHour = time.Now().Add(time.Hour)
		}
	}
}
