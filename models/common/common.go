package common

import (
	"html"
	"net/url"

	"github.com/merliot/dean"
)

// key: ssid; value: passphrase
type WifiAuth map[string]string

type Commoner interface {
	Load()
	SetWifiAuth(WifiAuth)
}

type Common struct {
	dean.Thing
	Targets      `json:"-"`
	WifiAuth     `json:"-"`
	DeployParams string
	commonOS
}

func New(id, model, name string, targets []string) dean.Thinger {
	println("NEW COMMON")
	c := &Common{}
	c.Thing = dean.NewThing(id, model, name)
	c.Targets = makeTargets(targets)
	c.WifiAuth = make(WifiAuth)
	c.commonOSInit()
	return c
}

func (c *Common) ParseDeployParams() url.Values {
	values, _ := url.ParseQuery(c.DeployParams)
	return values
}

func (c *Common) SetDeployParams(params string) {
       c.DeployParams = html.UnescapeString(params)
}

func (c *Common) SetWifiAuth(auth WifiAuth) {
	c.WifiAuth = auth
}
