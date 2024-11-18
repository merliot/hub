// Morningstar Prostar PWM array charge controller device.
//
// modbus ref: https://www.morningstarcorp.com/wp-content/uploads/technical-doc-prostar-modbus-specification-en.pdf

package prostar

import (
	"math"
	"strconv"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/io/modbus"
	"github.com/x448/float16"
)

const (
	regVerSw       = 0x0000
	regAdcIa       = 0x0011
	regAdcIl       = 0x0016
	regChargeState = 0x0021
	regLoadState   = 0x002E
	regVbMinDaily  = 0x0041
)

var (
	pollPeriod = 5 * time.Second
	ttyUSB0    = "/dev/ttyUSB0"
)

type Status string

type System struct {
	SWVersion     string
	BattVoltMulti uint16
}

type Controller struct {
	Amps float32
}

type Battery struct {
	Volts       float32
	SenseVolts  float32
	SlowVolts   float32
	SlowNetAmps float32
}

type Load struct {
	Volts float32
	Amps  float32
	State uint16
	Fault uint16
}

type Array struct {
	Volts float32
	Amps  float32
	State uint16
}

type Daily struct {
	MinBattVolts float32
	MaxBattVolts float32
	ChargeAh     float32
	LoadAh       float32
}

type prostar struct {
	TTY string
	Status
	System
	Controller
	Battery
	Load
	Array
	Daily
	modbus.Modbus
	nextHour time.Time
}

func NewModel() device.Devicer {
	return &prostar{TTY: ttyUSB0}
}

func (p *prostar) save(pkt *device.Packet) {
	pkt.Unmarshal(p).RouteUp()
}

func swap(b []byte) uint16 {
	return (uint16(b[0]) << 8) | uint16(b[1])
}

func unswap(v uint16) []byte {
	return []byte{byte(v >> 8), byte(v)}
}

func noswap(b []byte) uint16 {
	return (uint16(b[1]) << 8) | uint16(b[0])
}

func f16(b []byte) float32 {
	return float16.Float16(swap(b)).Float32()
}

func unf16(f32 float32) []byte {
	f16 := float16.Fromfloat32(f32)
	return []byte{byte(f16 >> 8), byte(f16 & 0xFF)}
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

// Round to 2 decimal places
func round2(num float32) float32 {
	rounded := math.Round(float64(num)*100) / 100
	return float32(rounded)
}

func (p *prostar) readDynamic(c *Controller, b *Battery, l *Load, s *Array) error {

	// FILTERED ADC

	regs, err := p.ReadRegisters(regAdcIa, 4)
	if err != nil {
		return err
	}

	s.Amps = round2(f16(regs[0:2]))
	b.Volts = round2(f16(regs[2:4]))
	s.Volts = round2(f16(regs[4:6]))
	l.Volts = round2(f16(regs[6:8]))

	regs, err = p.ReadRegisters(regAdcIl, 4)
	if err != nil {
		return err
	}

	l.Amps = round2(f16(regs[0:2]))
	b.SenseVolts = round2(f16(regs[2:4]))
	b.SlowVolts = round2(f16(regs[4:6]))
	b.SlowNetAmps = round2(f16(regs[6:8]))

	// CHARGER STATUS

	regs, err = p.ReadRegisters(regChargeState, 1)
	if err != nil {
		return err
	}

	s.State = swap(regs[0:2])

	// LOAD STATUS

	regs, err = p.ReadRegisters(regLoadState, 2)
	if err != nil {
		return err
	}

	l.State = swap(regs[0:2])
	l.Fault = swap(regs[2:4])

	return nil
}

var statusOK = Status("OK")

func (p *prostar) sendStatus(pkt *device.Packet, newStatus Status) {
	if newStatus == "EOF" {
		newStatus = "METERBUS DISCONNECTED"
	}
	if p.Status == newStatus {
		return
	}
	p.Status = newStatus
	pkt.SetPath("/update-status").Marshal(p.Status).RouteUp()
}

func (p *prostar) sendDynamic(pkt *device.Packet) {
	var controller Controller
	var battery Battery
	var load Load
	var array Array

	if err := p.readDynamic(&controller, &battery, &load, &array); err != nil {
		p.sendStatus(pkt, Status(err.Error()))
		return
	}

	// If anything has changed, send updates

	if controller != p.Controller {
		p.Controller = controller
		pkt.SetPath("/update-controller").Marshal(controller).RouteUp()
	}
	if battery != p.Battery {
		p.Battery = battery
		pkt.SetPath("/update-battery").Marshal(battery).RouteUp()
	}
	if load != p.Load {
		p.Load = load
		pkt.SetPath("/update-load").Marshal(load).RouteUp()
	}
	if array != p.Array {
		p.Array = array
		pkt.SetPath("/update-array").Marshal(array).RouteUp()
	}

	p.sendStatus(pkt, statusOK)
}

func (p *prostar) sendHourly(pkt *device.Packet) (err error) {
	var daily Daily

	if err = p.readDaily(&daily); err != nil {
		p.sendStatus(pkt, Status(err.Error()))
		return
	}

	// If anything has changed, send updates

	if daily != p.Daily {
		p.Daily = daily
		pkt.SetPath("/update-daily").Marshal(daily).RouteUp()
	}

	p.sendStatus(pkt, statusOK)
	return
}

func (p *prostar) readSystem(s *System) (err error) {
	var regs []byte

	for i := 0; i < 5; i++ {
		regs, err = p.ReadRegisters(regVerSw, 2)
		if err == nil {
			s.SWVersion = bcdToDecimal(swap(regs[0:2]))
			s.BattVoltMulti = noswap(regs[2:4])
			return
		}
		time.Sleep(time.Second)
	}

	return
}

func (p *prostar) readDaily(d *Daily) error {

	// LOGGER

	regs, err := p.ReadRegisters(regVbMinDaily, 4)
	if err != nil {
		return err
	}

	d.MinBattVolts = round2(f16(regs[0:2]))
	d.MaxBattVolts = round2(f16(regs[2:4]))
	d.ChargeAh = round2(f16(regs[4:6]))
	d.LoadAh = round2(f16(regs[6:8]))

	return nil
}

func (p *prostar) Setup() error {
	p.Modbus = modbus.New(newTransport(p.TTY))
	if err := p.readSystem(&p.System); err != nil {
		return err
	}
	if err := p.readDaily(&p.Daily); err != nil {
		return err
	}
	p.nextHour = time.Now().Add(time.Hour)
	return nil
}

func (p *prostar) Poll(pkt *device.Packet) {
	p.sendDynamic(pkt)
	if time.Now().After(p.nextHour) {
		p.sendHourly(pkt)
		p.nextHour = time.Now().Add(time.Hour)
	}
}

func (p *prostar) DemoSetup() error            { return p.Setup() }
func (p *prostar) DemoPoll(pkt *device.Packet) { p.Poll(pkt) }
