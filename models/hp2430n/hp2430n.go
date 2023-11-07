package hp2430n

import (
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/charge"
)

const (
	regBatteryVoltage  = 0x101
	regChargingCurrent = 0x102
	regTemperature     = 0x103
	regLoadVoltage     = 0x104
	regLoadCurrent     = 0x105
	regSolarVoltage    = 0x107
	regSolarCurrent    = 0x108
	regLoadInfo        = 0x120
)

const (
	batteryVoltage uint16 = iota
	chargingCurrent
	loadVoltage
	loadCurrent
	solarVoltage
	solarCurrent
	temperature
	fields
)

type record [fields]float32

type Hp2430n struct {
	*charge.Charge
	Seconds      []record
	Minutes      []record
	Hours        []record
	Days         []record
	LoadInfo     uint16
	targetStruct
}

var targets = []string{"x86-64", "rpi", "nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW HP2430N")
	h := &Hp2430n{}
	h.Charge = charge.New(id, model, name, targets).(*charge.Charge)
	h.Seconds = make([]record, 0)
	h.Minutes = make([]record, 0)
	h.Hours = make([]record, 0)
	h.Days = make([]record, 0)
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

func (h *Hp2430n) updateStatus(msg *dean.Msg) {
	msg.Unmarshal(h).Broadcast()
}

func (h *Hp2430n) saveRecord(recs *[]record, rec record, size int) {
	n := len(*recs)
	if n >= size {
		n = size - 1
	}
	// newest record at recs[0], oldest at recs[n-1]
	*recs = append([]record{rec}, (*recs)[:n]...)
}

type RecUpdateMsg struct {
	Path   string
	Record record
}

func (h *Hp2430n) updateRecord(recs *[]record, size int) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		var update RecUpdateMsg
		msg.Unmarshal(&update)
		h.saveRecord(recs, update.Record, size)
		msg.Broadcast()
	}
}

func (h *Hp2430n) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":         h.save,
		"get/state":     h.getState,
		"update/status": h.updateStatus,
		"update/second": h.updateRecord(&h.Seconds, 60),
		"update/minute": h.updateRecord(&h.Minutes, 60),
		"update/hour":   h.updateRecord(&h.Hours, 24),
		"update/day":    h.updateRecord(&h.Days, 365),
	}
}

func ave(recs []record) record {
	var rec record
	for j := 0; j < len(rec); j++ {
		var sum float32
		for i := 0; i < len(recs); i++ {
			sum += recs[i][j]
		}
		rec[j] = sum / float32(len(recs))
	}
	return rec
}

func (h *Hp2430n) nextRecord() (rec record) {
	rec[batteryVoltage]  = h.readVoltage(regBatteryVoltage)
	rec[chargingCurrent] = h.readCurrent(regChargingCurrent)
	rec[loadVoltage]     = h.readVoltage(regLoadVoltage)
	rec[loadCurrent]     = h.readCurrent(regLoadCurrent)
	rec[solarVoltage]    = h.readVoltage(regSolarVoltage)
	rec[solarCurrent]    = h.readCurrent(regSolarCurrent)
//	rec[temperature]     = h.readTemperature()
	return
}

func (h *Hp2430n) sendRecord(i *dean.Injector, tag string, rec record) {
	var msg dean.Msg
	var update = RecUpdateMsg{
		Path:   "update/" + tag,
		Record: rec,
	}
	i.Inject(msg.Marshal(&update))
}

type StatusMsg struct {
	Path     string
	LoadInfo uint16
}

func (h *Hp2430n) sendStatus(i *dean.Injector) {
	var msg dean.Msg
	var update = StatusMsg{
		Path: "update/status",
		LoadInfo: h.readLoadInfo(),
	}
	if update.LoadInfo != h.LoadInfo {
		h.LoadInfo = update.LoadInfo
		i.Inject(msg.Marshal(&update))
	}
}

func (h *Hp2430n) sample(i *dean.Injector) {
	ticker := time.NewTicker(time.Second)

	for {
		for day := 0; day < 365; day++ {
			for hr := 0; hr < 24; hr++ {
				for min := 0; min < 60; min++ {
					for sec := 0; sec < 60; sec++ {
						select {
						case <-ticker.C:
							h.sendStatus(i)
							h.sendRecord(i, "second", h.nextRecord())
						}
					}
					h.sendRecord(i, "minute", ave(h.Seconds[:60]))
				}
				h.sendRecord(i, "hour", ave(h.Minutes[:60]))
			}
			h.sendRecord(i, "day", ave(h.Hours[:24]))
		}
	}
}
