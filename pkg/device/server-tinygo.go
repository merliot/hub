//go:build tinygo

package device

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/merliot/hub/pkg/tinynet"
)

type server struct {
	maker          Maker
	devices        deviceMap
	uplinks        uplinkMap
	sessions       sessionMap
	packetHandlers PacketHandlers
	root           *device
	port           int
	logLevel       string
}

var paramsMem = []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

func NewServer(maker Maker) *server {
	return &server{maker: maker}
}

func (s *server) devicesOnline(l linker) {

	var pkt = &Packet{}
	pkt.SetDst(s.root.Id).SetPath("online").Marshal(s.root.State)

	s.logInfo("Sending", "pkt", pkt)
	l.Send(pkt)
}

func (s *server) Run() {
	var params uf2ParamsBlock

	// wait a bit for serial
	time.Sleep(2 * time.Second)

	end := bytes.IndexByte(paramsMem, 0)
	if err := json.Unmarshal(paramsMem[:end], &params); err != nil {
		panic(err)
	}

	if params.MagicStart != uf2Magic || params.MagicEnd != uf2Magic {
		panic("UF2 magic incorrect")
	}

	s.logLevel = params.LogLevel

	tinynet.NetConnect(params.Ssid, params.Passphrase)

	s.root = &device{
		Id:           params.Id,
		Model:        params.Model,
		Name:         params.Name,
		DeployParams: params.DeployParams,
		model:        &Model{Maker: s.maker},
		server:       s,
	}

	if err := s.build(s.root, 0); err != nil {
		panic(err)
	}

	s.root.set(flagOnline | flagMetal)

	if err := s.root.Setup(); err != nil {
		panic(err)
	}

	s.dialParents(params.DialURLs, params.User, params.Passwd)

	var pkt = &Packet{Dst: s.root.Id}

	ticker := time.NewTicker(s.root.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.root.stateMu.Lock()
			s.root.Poll(pkt)
			s.root.stateMu.Unlock()
		}
	}
}

func (s *server) routeDown(pkt *Packet) error {
	s.logDebug("routeDown", "pkt", pkt)
	s.root.handle(pkt)
	return nil
}
