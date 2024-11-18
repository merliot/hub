package qrcode

import (
	"github.com/merliot/hub/pkg/device"
)

func NewModel() device.Devicer {
	return &qrcode{}
}

func (q *qrcode) DemoSetup() error {
	q.Content = "https://merliot.io"
	return nil
}

func (q *qrcode) Poll(pkt *device.Packet)     {}
func (q *qrcode) DemoPoll(pkt *device.Packet) {}
