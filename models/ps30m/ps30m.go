package ps30m

import (
	"embed"
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
	"github.com/simonvetter/modbus"
	"github.com/x448/float16"
)

//go:embed *
var fs embed.FS

const (
	AdcIa       = 0x0011
	AdcVbterm   = 0x0012
	AdcIl       = 0x0016
	ChargeState = 0x0021
	LoadState   = 0x002E
)

type record [3]float32

type Ps30m struct {
	*common.Common
	ChargeStatus uint16
	LoadStatus   uint16
	Seconds      []record
	Minutes      []record
	Hours        []record
	Days         []record
	client       *modbus.ModbusClient
	demo         bool
	templates    *template.Template
}

func New(id, model, name string) dean.Thinger {
	println("NEW PS30M")
	p := &Ps30m{}
	p.Common = common.New(id, model, name).(*common.Common)
	p.Seconds = make([]record, 0)
	p.Minutes = make([]record, 0)
	p.Hours = make([]record, 0)
	p.Days = make([]record, 0)
	p.CompositeFs.AddFS(fs)
	p.templates = p.CompositeFs.ParseFS("template/*")
	return p
}

func (p *Ps30m) save(msg *dean.Msg) {
	msg.Unmarshal(p).Broadcast()
}

func (p *Ps30m) getState(msg *dean.Msg) {
	p.Path = "state"
	msg.Marshal(p).Reply()
}

func (p *Ps30m) updateStatus(msg *dean.Msg) {
	msg.Unmarshal(p).Broadcast()
}

func (p *Ps30m) saveRecord(recs *[]record, rec record, size int) {
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

func (p *Ps30m) updateRecord(recs *[]record, size int) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		var update RecUpdateMsg
		msg.Unmarshal(&update)
		p.saveRecord(recs, update.Record, size)
		msg.Broadcast()
	}
}

func (p *Ps30m) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":         p.save,
		"get/state":     p.getState,
		"update/status": p.updateStatus,
		"update/second": p.updateRecord(&p.Seconds, 60),
		"update/minute": p.updateRecord(&p.Minutes, 60),
		"update/hour":   p.updateRecord(&p.Hours, 24),
		"update/day":    p.updateRecord(&p.Days, 365),
	}
}

func (p *Ps30m) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/deploy?target={target}\n"))
	w.Write([]byte("/state\n"))
}

func (p *Ps30m) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "api":
		p.api(w, r)
	case "state":
		common.ShowState(p.templates, w, p)
	default:
		p.API(p.templates, w, r)
	}
}

type RegsUpdateMsg struct {
	Path string
	Regs map[uint16]any // keyed by addr
}

func (p *Ps30m) readRegF32(addr uint16) float32 {
	value, _ := p.client.ReadRegister(addr, modbus.HOLDING_REGISTER)
	return float16.Float16(value).Float32()
}

func (p *Ps30m) readRegU16(addr uint16) uint16 {
	value, _ := p.client.ReadRegister(addr, modbus.HOLDING_REGISTER)
	return value
}

func (p *Ps30m) Demo() {
	p.demo = true
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

var sun = [...]float32{
	0.0, 0.0, 0.0, 0.0,
	0.0, 0.0, 1.0, 1.5,
	2.5, 4.0, 7.0, 9.0,
	12.0, 13.0, 13.0, 12.0,
	9.0, 7.0, 4.0, 2.5,
	1.0, 0.0, 0.0, 0.0,
}

func (p *Ps30m) nextRecord() (rec record) {
	if p.demo {
		hour := time.Now().Hour()
		rec[0] = sun[hour] + rand.Float32()
		rec[1] = 13.0 + rand.Float32()
		rec[2] = 3 + rand.Float32()
		return
	}

	rec[0] = p.readRegF32(AdcIa)
	rec[1] = p.readRegF32(AdcVbterm)
	rec[2] = p.readRegF32(AdcIl)
	return
}

func (p *Ps30m) sendRecord(i *dean.Injector, tag string, rec record) {
	var msg dean.Msg
	var update = RecUpdateMsg{
		Path:   "update/" + tag,
		Record: rec,
	}
	i.Inject(msg.Marshal(&update))
}

type StatusMsg struct {
	Path        string
	ChargeState uint16
	LoadState   uint16
}

func (p *Ps30m) sendStatus(i *dean.Injector) {
	var msg dean.Msg
	var update = StatusMsg{Path: "update/status"}
	if p.demo {
		update.ChargeState = 1
		update.LoadState = 3
	} else {
		update.ChargeState = p.readRegU16(ChargeState)
		update.LoadState = p.readRegU16(LoadState)
	}
	i.Inject(msg.Marshal(&update))
}

func (p *Ps30m) sample(i *dean.Injector) {
	ticker := time.NewTicker(time.Second)

	for {
		for day := 0; day < 365; day++ {
			for hr := 0; hr < 24; hr++ {
				for min := 0; min < 60; min++ {
					for sec := 0; sec < 60; sec++ {
						select {
						case <-ticker.C:
							p.sendStatus(i)
							p.sendRecord(i, "second", p.nextRecord())
						}
					}
					p.sendRecord(i, "minute", ave(p.Seconds[:60]))
				}
				p.sendRecord(i, "hour", ave(p.Minutes[:60]))
			}
			p.sendRecord(i, "day", ave(p.Hours[:24]))
		}
	}
}

func (p *Ps30m) Run(i *dean.Injector) {
	const serial = "rtu:///dev/ttyUSB0"
	var err error

	if !p.demo {
		p.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:      serial,
			Speed:    9600,
			DataBits: 8,
			Parity:   modbus.PARITY_NONE,
			StopBits: 2,
			Timeout:  300 * time.Millisecond,
		})
		if err != nil {
			println("Create modbus client failed:", err.Error())
			return
		}

		if err = p.client.Open(); err != nil {
			println("Open modbus client at", serial, "failed:", err.Error())
			return
		}
	}

	p.sample(i)
}
