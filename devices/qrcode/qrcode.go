package qrcode

import (
	"github.com/merliot/hub/pkg/device"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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

func (q *qrcode) Detail() Node {
	return Div(
		Class("flex flex-col items-center"),
		Attr("id", "qrcode-detail"),
		Div(
			Class("flex flex-row items-center"),
			Attr("id", "qrcode-edit-content"),
			Span(Class("m-4"), Text(q.Content)),
			// TODO: Add edit button logic and lock state if needed
			Img(Class("icon"), Src("/images/edit.svg")),
		),
		Div(
			Class("flex flex-row justify-center"),
			Div(
				Class("m-8"),
				// TODO: Replace with actual PNG generation for QR code
				Img(Src("/model/qrcode/images/qr-placeholder.png")),
			),
		),
	)
}

func (q *qrcode) Overview() Node {
	return Div(
		Class("flex flex-col items-center"),
		Attr("id", "qrcode-overview"),
		// TODO: Replace with actual PNG generation for QR code
		Img(Class("mr-4"), Src("/model/qrcode/images/qr-placeholder.png")),
		Span(Text(q.Content)),
	)
}
