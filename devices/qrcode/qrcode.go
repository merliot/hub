package qrcode

import (
	"github.com/merliot/hub"
)

func NewModel() hub.Devicer {
	return &qrcode{}
}

func (q *qrcode) Poll(pkt *hub.Packet)     {}
func (q *qrcode) DemoSetup() error         { return nil }
func (q *qrcode) DemoPoll(pkt *hub.Packet) {}
