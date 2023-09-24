package common

import (
	"html"
	"net/url"

	"github.com/merliot/dean"
)

type Wifiver interface {
	SetWifiAuth(ssid, passphrase string)
}

type Common struct {
	dean.Thing
	Targets `json:"-"`
	DeployParams string
	Demo         bool `json:"-"`
	ssid         string
	passphrase   string
	commonOS
}

func New(id, model, name string, targets []string) dean.Thinger {
	println("NEW COMMON")
	c := &Common{}
	c.Thing = dean.NewThing(id, model, name)
	c.Targets = makeTargets(targets)
	c.commonOSInit()
	return c
}

func (c *Common) ParseDeployParams() url.Values {
	unescaped := html.UnescapeString(c.DeployParams)
	values, _ := url.ParseQuery(unescaped)
	return values
}

func (c *Common) SetWifiAuth(ssid, passphrase string) {
	c.ssid = ssid
	c.passphrase = passphrase
}
