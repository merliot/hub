package qrcode

import (
	"fmt"

	"github.com/merliot/hub"
	qr "github.com/skip2/go-qrcode"
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
	}
}

func (q *qrcode) GetHandlers() hub.Handlers {
	return hub.Handlers{
		"/update": &hub.Handler[qrcode]{q.update},
	}
}

func (q *qrcode) update(pkt *hub.Packet) {
}

func (q *qrcode) Setup() error {
	code, err := qr.New(q.Content, qr.High)
	if err != nil {
		return err
	}
	fmt.Println(code.ToString(true))
	return nil
}

func (q *qrcode) Poll(pkt *hub.Packet)     {}
func (q *qrcode) DemoSetup() error         { return nil }
func (q *qrcode) DemoPoll(pkt *hub.Packet) {}
