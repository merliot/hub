package ps30m

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"math/rand"
	"strconv"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
	"github.com/simonvetter/modbus"
	"github.com/x448/float16"
)

//go:embed css html js
var fs embed.FS
var tmpls = template.Must(template.ParseFS(fs, "html/*"))

const (
	AdcIa = 0x0011
	AdcVbterm = 0x0012
	AdcIl = 0x0016
	ChargeState = 0x0021
	LoadState = 0x002E
)

type record [3]float32

type Ps30m struct {
	*common.Common
	Regs map[uint16] any // keyed by addr
	Seconds []record
	Minutes []record
	Hours []record
	Days []record
	client *modbus.ModbusClient
	demo bool
}

func New(id, model, name string) dean.Thinger {
	println("NEW PS30M")
	return &Ps30m{
		Common: common.New(id, model, name).(*common.Common),
		Seconds: make([]record, 0),
		Minutes: make([]record, 0),
		Hours: make([]record, 0),
		Days: make([]record, 0),
	}
}

func (p *Ps30m) save(msg *dean.Msg) {
	msg.Unmarshal(p)
}

func (p *Ps30m) getState(msg *dean.Msg) {
	p.Path = "state"
	msg.Marshal(p).Reply()
}

func (p *Ps30m) regsUpdate(msg *dean.Msg) {
	msg.Unmarshal(p).Broadcast()
}

func (p *Ps30m) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     p.save,
		"get/state": p.getState,
		"regs/update": p.regsUpdate,
	}
}

func (p *Ps30m) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/" + p.Id() + "/api\n"))
	w.Write([]byte("/" + p.Id() + "/readreg?addr={addr}&type={type}\n"))
	w.Write([]byte("\ttype = 0 holding register\n"))
	w.Write([]byte("\ttype = 1 input register\n"))
}

type reg struct {
	Addr uint16
	Value any
	Err error
}

func (p *Ps30m) readreg(w http.ResponseWriter, r *http.Request) {
	var reg reg
	var regaddr int64

	regaddr, reg.Err = strconv.ParseInt(r.URL.Query().Get("addr"), 0, 16)
	reg.Addr = uint16(regaddr)
	reg.Value = p.Regs[reg.Addr]

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	enc.Encode(reg)
}

func (p *Ps30m) API(fs embed.FS, tmpls *template.Template, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "api":
		p.api(w, r)
	case "readreg":
		p.readreg(w, r)
	default:
		p.Common.API(fs, tmpls, w, r)
	}
}

func (p *Ps30m) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.API(fs, tmpls, w, r)
}

type RegsUpdateMsg struct {
	Path string
	Regs map[uint16] any // keyed by addr
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

var sun = [...]float32{
	0.0, 0.0, 0.0, 0.0,
	0.0, 0.0, 1.0, 1.5,
	2.5, 4.0, 7.0, 9.0,
	12.0, 13.0, 13.0, 12.0,
	9.0, 7.0, 4.0, 2.5,
	1.0, 0.0, 0.0, 0.0,
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

func (p *Ps30m) send(tag string, rec record) {
	println(tag)
}

func (p *Ps30m) store(recs *[]record, rec record, size int, tag string) {
	*recs = append(*recs, rec)
	if len(*recs) > size {
		*recs = (*recs)[1:]
	}
	p.send(tag, (*recs)[len(*recs)-1])
}

func (p *Ps30m) sample(next func () record) {
	ticker := time.NewTicker(time.Second)

	for {
		for day := 0; day < 365; day++ {
			for hr := 0; hr < 24; hr++ {
				for min := 0; min < 60; min++ {
					for sec := 0; sec < 60; sec++ {
						select {
						case <-ticker.C:
							p.store(&p.Seconds, next(), 60, "second")
						}
					}
					p.store(&p.Minutes, ave(p.Seconds), 60, "minute")
				}
				p.store(&p.Hours, ave(p.Minutes), 24, "hour")
			}
			p.store(&p.Days, ave(p.Hours), 365, "day")
		}
	}
}

func (p *Ps30m) sampleDemo(i *dean.Injector, msg *dean.Msg, update *RegsUpdateMsg) record {
	var rec record

	hour := time.Now().Hour()
	rec[0] = sun[hour] + rand.Float32()
	rec[1] = 13.0 + rand.Float32()
	rec[2] = 3 + rand.Float32()

	update.Regs[AdcIa] =  rec[0]
	update.Regs[AdcVbterm] = rec[1]
	update.Regs[AdcIl] = rec[2]
	update.Regs[ChargeState] = 0
	update.Regs[LoadState] = 0

	i.Inject(msg.Marshal(update))

	return rec
}

func (p *Ps30m) runDemo(i *dean.Injector, msg *dean.Msg, update *RegsUpdateMsg) {
	p.sample(func() record {
		return p.sampleDemo(i, msg, update)
	})
}

func (p *Ps30m) sampleRun(i *dean.Injector, msg *dean.Msg, update *RegsUpdateMsg) record {
	var rec record

	rec[0] = p.readRegF32(AdcIa)
	rec[1] = p.readRegF32(AdcVbterm)
	rec[2] = p.readRegF32(AdcIl)

	update.Regs[AdcIa] = rec[0]
	update.Regs[AdcVbterm] = rec[1]
	update.Regs[AdcIl] = rec[2]
	update.Regs[ChargeState] = p.readRegU16(ChargeState)
	update.Regs[LoadState] = p.readRegU16(LoadState)
	i.Inject(msg.Marshal(&update))

	return rec
}

func (p *Ps30m) Run(i *dean.Injector) {
	var err error
	var msg dean.Msg
	var update = RegsUpdateMsg{
		Path: "regs/update",
		Regs: make(map[uint16] any),
	}

	if p.demo {
		p.runDemo(i, &msg, &update)
		return
	}

	p.client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:      "rtu:///dev/ttyUSB0",
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 2,
		Timeout:  300 * time.Millisecond,
	})
	if err != nil {
		panic(err.Error())
	}

	if err = p.client.Open(); err != nil {
		panic(err.Error())
	}

	p.sample(func() record {
		return p.sampleRun(i, &msg, &update)
	})
}
