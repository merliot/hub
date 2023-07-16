package ps30m

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
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

type Ps30m struct {
	*common.Common
	client *modbus.ModbusClient
	Regs map[uint16] any // keyed by addr
}

func New(id, model, name string) dean.Thinger {
	println("NEW PS30M")
	return &Ps30m{
		Common: common.New(id, model, name).(*common.Common),
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

func (p *Ps30m) Run(i *dean.Injector) {
	var err error
	var msg dean.Msg
	var update = RegsUpdateMsg{
		Path: "regs/update",
		Regs: make(map[uint16] any),
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

	for {
		update.Regs[AdcIa] = p.readRegF32(AdcIa)
		update.Regs[AdcVbterm] = p.readRegF32(AdcVbterm)
		update.Regs[AdcIl] = p.readRegF32(AdcIl)
		update.Regs[ChargeState] = p.readRegU16(ChargeState)
		update.Regs[LoadState] = p.readRegU16(LoadState)
		i.Inject(msg.Marshal(&update))
		time.Sleep(time.Second)
	}
}
