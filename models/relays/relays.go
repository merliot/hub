package relays

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type Relays struct {
	*common.Common
	Relays    [4]Relay
	relaysOS
}

type MsgClick struct {
	dean.ThingMsg
	Relay int
	State bool
}

var targets = []string{"x86-64", "rpi", "nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW RELAYS")
	r := &Relays{}
	r.Common = common.New(id, model, name, targets).(*common.Common)
	r.relaysOSNew()
	return r
}

func (r *Relays) save(msg *dean.Msg) {
	msg.Unmarshal(r).Broadcast()
}

func (r *Relays) getState(msg *dean.Msg) {
	r.Path = "state"
	msg.Marshal(r).Reply()
}

func (r *Relays) click(msg *dean.Msg) {
	var msgClick MsgClick
	msg.Unmarshal(&msgClick)
	relay := &r.Relays[msgClick.Relay]
	relay.State = msgClick.State
	if r.IsMetal() {
		if msgClick.State {
			relay.On()
		} else {
			relay.Off()
		}
	}
	msg.Broadcast()
}

func (r *Relays) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     r.save,
		"get/state": r.getState,
		"click":     r.click,
	}
}

func (r *Relays) setRelay(num int, name, gpio string) {
	relay := &r.Relays[num]
	if name == "" {
		name = fmt.Sprintf("Relay #%d", num+1)
	}
	relay.Name = name
	relay.Gpio = gpio
}

func firstValue(values url.Values, key string) string {
	if v, ok := values[key]; ok {
		return v[0]
	}
	return ""
}

func (r *Relays) parseParams() {
	values := r.ParseDeployParams()
	r.Demo = (firstValue(values, "demo") == "on")
	for i, _ := range r.Relays {
		num := strconv.Itoa(i + 1)
		name := firstValue(values, "relay" + num)
		gpio := firstValue(values, "gpio" + num)
		r.setRelay(i, name, gpio)
	}
}

func (r *Relays) Run(i *dean.Injector) {
	r.parseParams()
	if r.Demo {
		r.runDemo(i)
		return
	}
	r.runOS(i)
}
