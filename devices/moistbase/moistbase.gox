package moistbase

import (
	"embed"
	"time"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go template
var fs embed.FS

type moistbase struct {
}

func NewModel() device.Devicer {
	return &moistbase{}
}

func (m *moistbase) GetConfig() device.Config {
	return device.Config{
		Model:      "moistbase",
		Parents:    []string{"hub"},
		Flags:      device.FlagProgenitive,
		State:      m,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi"},
		PollPeriod: time.Second,
		BgColor:    "moonlit-violet",
		FgColor:    "black",
		PacketHandlers: device.PacketHandlers{
			"/created":   &device.PacketHandler[device.MsgCreated]{m.created},
			"/destroyed": &device.PacketHandler[device.MsgDestroyed]{m.destroyed},
		},
	}
}

func (m *moistbase) created(pkt *device.Packet) {
	var msg device.MsgCreated
	pkt.Unmarshal(&msg)
	device.LogInfo("Created", "msg", msg)
	pkt.BroadcastUp()
}

func (m *moistbase) destroyed(pkt *device.Packet) {
	var msg device.MsgDestroyed
	pkt.Unmarshal(&msg)
	device.LogInfo("Destroyed", "msg", msg)
	pkt.BroadcastUp()
}

func (m *moistbase) Setup() error                { return nil }
func (m *moistbase) Poll(pkt *device.Packet)     {}
func (m *moistbase) DemoSetup() error            { return m.Setup() }
func (m *moistbase) DemoPoll(pkt *device.Packet) { m.Poll(pkt) }
