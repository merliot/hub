package common

import (
	"html"
	"net/url"

	"github.com/merliot/dean"
)

type Devicer interface {
	Load()
	SetWifiAuth(ssid, passphrase string)
	SetDeployParams(params string)
}

type Common struct {
	dean.Thing
	Targets      `json:"-"`
	DeployParams string `json:"-"`
	Demo         bool   `json:"-"`
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

func (c *Common) SetWifiAuth(ssid, passphrase string) {
	c.ssid = ssid
	c.passphrase = passphrase
}

func (c *Common) GetWifiAuth() (ssid, passphrase string) {
	return c.ssid, c.passphrase
}

func (c *Common) ParseDeployParams() url.Values {
	values, _ := url.ParseQuery(c.DeployParams)
	return values
}

func (c *Common) SetDeployParams(params string) {
	c.DeployParams = html.UnescapeString(params)
}
