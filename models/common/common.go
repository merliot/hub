package common

import (
	"encoding/json"
	"html"
	"net/url"
	"os"

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

func (c *Common) Load() {
	bytes, err := os.ReadFile("devs/" + c.Id + ".json")
	if err == nil {
		json.Unmarshal(bytes, &c.DeployParams)
	}
}

func (c *Common) Save() {
	bytes, err := json.MarshalIndent(c.DeployParams, "", "\t")
	if err == nil {
		os.WriteFile("devs/"+c.Id+".json", bytes, 0600)
	}
}
