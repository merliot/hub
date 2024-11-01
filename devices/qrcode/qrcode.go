package qrcode

import (
	"github.com/merliot/hub"
)

type qrcode struct {
	Content string
}

func NewModel() hub.Devicer {
	return &qrcode{}
}

func (q *qrcode) GetConfig() hub.Config {
	return hub.Config{
		Model:   "qrcode",
		State:   q,
		FS:      &fs,
		Targets: []string{"wioterminal"},
		BgColor: "butterscotch",
		FgColor: "black",
		APIs: hub.APIs{
			"POST /generate": q.generate,
		},
	}
}

func (q *qrcode) Setup() error             { return nil }
func (q *qrcode) Poll(pkt *hub.Packet)     {}
func (q *qrcode) DemoSetup() error         { return nil }
func (q *qrcode) DemoPoll(pkt *hub.Packet) {}
