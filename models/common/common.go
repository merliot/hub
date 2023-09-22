package common

import (
	"github.com/merliot/dean"
)

type Wifiver interface {
	SetWifiAuth(ssid, passphrase string)
}

type Common struct {
	dean.Thing
	Targets `json:"-"`
	commonOS
	ssid       string
	passphrase string
	deployURL  string
}

func New(id, model, name string, targets []string) dean.Thinger {
	println("NEW COMMON")
	c := &Common{}
	c.Thing = dean.NewThing(id, model, name)
	c.Targets = makeTargets(targets)
	c.commonOSInit()
	return c
}

func (c *Common) SetWifiAuth(ssid, passphrase string) {
	c.ssid = ssid
	c.passphrase = passphrase
}

func (c *Common) SetDeployURL(url string) {
	c.deployURL = url
}
