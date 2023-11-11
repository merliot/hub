package hp2430n

import (
	"fmt"
	"strings"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

const (
	regMaxVoltage      = 0x000A
	regBatteryCapacity = 0x0100
	regLoadInfo        = 0x0120
	regLoadCmd         = 0x010A
	regOpDays          = 0x0115
	regAlarm           = 0x0121
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
}

type Controller struct {
	Temp   uint8    // deg C
	Alarms []string // faults and warnings
}

func (c Controller) isDiff(other Controller) bool {
	if c.Temp != other.Temp {
		return true
	}
	if len(c.Alarms) != len(other.Alarms) {
		return true
	}
	for i, v := range c.Alarms {
		if v != other.Alarms[i] {
		    return true
		}
	}
	return false
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

type Daily struct {
	BattMinVolts      float32
	BattMaxVolts      float32
	ChargeMaxAmps     float32
	DischargeMaxAmps  float32
	ChargeMaxWatts    uint16
	DischargeMaxWatts uint16
	ChargeAmpHrs      uint16
	DischargeAmpHrs   uint16
	GenPowerWatts     uint16
	ConPowerWatts     uint16
}

type Historical struct {
	OpDays          uint16
	OverDischarges  uint16
	FullCharges     uint16
	ChargeAmpHrs    uint32
	DischargeAmpHrs uint32
	GenPowerWatts   uint32
	ConPowerWatts   uint32
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

type msgDaily struct {
	Path  string
	Daily Daily
}

type msgHistorical struct {
	Path       string
	Historical Historical
}

type Hp2430n struct {
	*common.Common
	System     System
	Controller Controller
	Battery    Battery
	LoadInfo   LoadInfo
	Solar      Solar
	Daily      Daily
	Historical Historical
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

func (h *Hp2430n) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":             h.save,
		"get/state":         h.getState,
		"update/controller": h.save,
		"update/battery":    h.save,
		"update/load":       h.save,
		"update/solar":      h.save,
		"update/dialy":      h.save,
		"update/historical": h.save,
	}
}

func version(b []byte) string {
	return fmt.Sprintf("%02d.%02d.%02d", b[1], b[2], b[3])
}

func serial(b []byte) string {
	return fmt.Sprintf("%02X%02X-%02X%02X", b[0], b[1], b[2], b[3])
}

func swap(b []byte) uint16 {
	return (uint16(b[0]) << 8) | uint16(b[1])
}

func swap4(b []byte) uint32 {
	return (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3])
}

func volts(b []byte) float32 {
	return float32(swap(b)) * 0.1
}

func amps(b []byte) float32 {
	return float32(swap(b)) * 0.01
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

var alarms = []string{
	"Battery over-discharge",
	"Battery over-voltage",
	"Battery under-voltage",
	"Load short circuit",
	"Load over-power or load over-current",
	"Controller temperature too high",
	"Battery high temperature protection (temperature higher than the upper discharge limit); prohibit charging",
	"Solar input over-power",
	"(reserved)",
	"Solar input side over-voltage",
	"(reserved)",
	"Solar panel working point over-voltage",
	"Solar panel reverse connected",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"(reserved)",
	"Power main supply",
	"OO battery detected (SLD)",
	"Battery high temperature protection (temperature higher than the upper discharge limit); prohibit discharging",
	"Battery low temperature protection (temperature lower than the lower discharge limit); prohibit discharging",
	"Over-charge protection; stop charging",
	"Battery low temperature protection (temperature is lower than the lower limit of charging; stop charging",
	"Battery reverse connected",
	"Capacitor over-voltage (reserved)",
	"Induction probe damaged (street light)",
	"Load open-circuit (street light)",
}

func parseAlarms(b []byte) (active []string) {
	value := swap4(b)
	for i := 0; i < 32; i++ {
		// Check if the bit is set
		if value&(1<<i) != 0 {
			// Add corresponding alarm
			active = append(active, alarms[i])
		}
	}
	return
}

func (h *Hp2430n) readDynamic(c *Controller, b *Battery, l *LoadInfo, s *Solar) error {

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

	// Controller alarm information
	regs, err = h.readRegisters(regAlarm, 2)
	if err != nil {
		return err
	}
	c.Alarms = parseAlarms(regs)

	return nil
}

func (h *Hp2430n) readDaily(d *Daily) error {

	// Current Day Info (22 bytes)
	regs, err := h.readRegisters(regLoadCmd, 11)
	if err != nil {
		return err
	}
	// skip load cmd regs[0:2]
	d.BattMinVolts = volts(regs[2:4])
	d.BattMaxVolts = volts(regs[4:6])
	d.ChargeMaxAmps = amps(regs[6:8])
	d.DischargeMaxAmps = amps(regs[8:10])
	d.ChargeMaxWatts = swap(regs[10:12])
	d.DischargeMaxWatts = swap(regs[12:14])
	d.ChargeAmpHrs = swap(regs[14:16])
	d.DischargeAmpHrs = swap(regs[16:18])
	d.GenPowerWatts = swap(regs[18:20])
	d.ConPowerWatts = swap(regs[20:22])
	return nil
}

func (h *Hp2430n) readHistorical(d *Historical) error {

	// Historical Info (22 bytes)
	regs, err := h.readRegisters(regOpDays, 11)
	if err != nil {
		return err
	}
	d.OpDays = swap(regs[0:2])
	d.OverDischarges = swap(regs[2:4])
	d.FullCharges = swap(regs[4:6])
	d.ChargeAmpHrs = swap4(regs[6:10])
	d.DischargeAmpHrs = swap4(regs[10:14])
	d.GenPowerWatts = swap4(regs[14:18])
	d.ConPowerWatts = swap4(regs[18:22])
	return nil
}

func (h *Hp2430n) Run(i *dean.Injector) {

	var msg dean.Msg
	var controller = msgController{Path: "update/controller"}
	var battery = msgBattery{Path: "update/battery"}
	var loadInfo = msgLoadInfo{Path: "update/load"}
	var solar = msgSolar{Path: "update/solar"}
	var daily = msgDaily{Path: "update/daily"}
	var historical = msgHistorical{Path: "update/historical"}

	// Read initial values

	h.Lock()

	if err := h.readSystem(&h.System); err != nil {
		println(err.Error())
	}
	if err := h.readDynamic(&h.Controller, &h.Battery, &h.LoadInfo, &h.Solar); err != nil {
		println(err.Error())
	}
	if err := h.readDaily(&h.Daily); err != nil {
		println(err.Error())
	}
	if err := h.readHistorical(&h.Historical); err != nil {
		println(err.Error())
	}

	h.Unlock()

	// Copy the initial values into the msgs

	controller.Controller = h.Controller
	battery.Battery = h.Battery
	loadInfo.LoadInfo = h.LoadInfo
	solar.Solar = h.Solar
	daily.Daily = h.Daily
	historical.Historical = h.Historical

	nextHour := time.Now().Add(time.Hour)
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {

		// Every tick, check if dynamic values have changed

		err := h.readDynamic(&controller.Controller, &battery.Battery, &loadInfo.LoadInfo, &solar.Solar)
		if err != nil {
			println(err.Error())
		}
		if controller.Controller.isDiff(h.Controller) {
			i.Inject(msg.Marshal(controller))
		}
		if battery.Battery != h.Battery {
			i.Inject(msg.Marshal(battery))
		}
		if loadInfo.LoadInfo != h.LoadInfo {
			i.Inject(msg.Marshal(loadInfo))
		}
		if solar.Solar != h.Solar {
			i.Inject(msg.Marshal(solar))
		}

		// Every hour, check if daily or historical values have changed

		if time.Now().After(nextHour) {
			err := h.readDaily(&daily.Daily)
			if err != nil {
				println(err.Error())
			}
			if daily.Daily != h.Daily {
				i.Inject(msg.Marshal(daily))
			}
			err = h.readHistorical(&historical.Historical)
			if err != nil {
				println(err.Error())
			}
			if historical.Historical != h.Historical {
				i.Inject(msg.Marshal(historical))
			}
			nextHour = time.Now().Add(time.Hour)
		}
	}
}
