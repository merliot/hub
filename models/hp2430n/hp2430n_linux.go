//go:build !tinygo

package hp2430n

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
	"github.com/simonvetter/modbus"
)

//go:embed *
var fs embed.FS

type targetStruct struct {
	templates *template.Template
	client    *modbus.ModbusClient
}

func (h *Hp2430n) targetNew() {
	h.CompositeFs.AddFS(fs)
	h.templates = h.CompositeFs.ParseFS("template/*")
}

func (h *Hp2430n) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "state":
		common.ShowState(h.templates, w, h)
	default:
		h.API(h.templates, w, r)
	}
}

func (h *Hp2430n) readRegU16(addr uint16) uint16 {
	value, _ := h.client.ReadRegister(addr, modbus.HOLDING_REGISTER)
	return value
}

func (h *Hp2430n) readVoltage(reg uint16) float32 {
	return float32(h.readRegU16(reg)) * 0.1 // Volts
}

func (h *Hp2430n) readCurrent(reg uint16) float32 {
	return float32(h.readRegU16(reg)) * 0.01 // Amps
}

func (h *Hp2430n) readLoadInfo() uint16 {
	return h.readRegU16(regLoadInfo)
}

func (h *Hp2430n) Run(i *dean.Injector) {
	const serial = "rtu:///dev/ttyUSB0"
	var err error

	h.client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:      serial,
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 1,
		Timeout:  300 * time.Millisecond,
	})
	if err != nil {
		println("Create modbus client failed:", err.Error())
		return
	}

	if err = h.client.Open(); err != nil {
		println("Open modbus client at", serial, "failed:", err.Error())
		return
	}

	h.sample(i)
}
