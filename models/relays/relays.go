package relays

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed *
var fs embed.FS

type Relay struct {
	Name   string
	Gpio   string
	State  bool
	driver *gpio.RelayDriver
}

func (r *Relay) Start() {
	if r.driver != nil {
		r.driver.Start()
	}
}

func (r *Relay) On() {
	if r.driver != nil {
		r.driver.On()
	}
}

func (r *Relay) Off() {
	if r.driver != nil {
		r.driver.Off()
	}
}

type Relays struct {
	*common.Common
	Relays    [4]Relay
	adaptor   *raspi.Adaptor
	templates *template.Template
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
	r.CompositeFs.AddFS(fs)
	r.adaptor = raspi.NewAdaptor()
	r.templates = r.CompositeFs.ParseFS("template/*")
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

func (r *Relays) api(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/deploy?target={target}\n"))
	w.Write([]byte("/state\n"))
}

func (r *Relays) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch strings.TrimPrefix(req.URL.Path, "/") {
	case "api":
		r.api(w, req)
	case "state":
		common.ShowState(r.templates, w, r)
	default:
		r.Common.API(r.templates, w, req)
	}
}

func (r *Relays) setRelay(num int, name, pin string) {
	relay := &r.Relays[num]
	if name == "" {
		name = fmt.Sprintf("Relay #%d", num+1)
	}
	relay.Name = name
	relay.Gpio = pin
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
		pin := firstValue(values, "gpio" + num)
		r.setRelay(i, name, pin)
	}
}

func (r *Relays) Run(i *dean.Injector) {

	// Fail safe by turning off relays
	failSafe := func () {
		if recover() != nil {
			for _, relay := range r.Relays {
				relay.Off()
			}
		}
	}
	defer failSafe()

	r.parseParams()

	r.adaptor.Connect()

	for i, _ := range r.Relays {
		relay := &r.Relays[i]
		if r.Demo || relay.Gpio == "" {
			continue
		}
		relay.driver = gpio.NewRelayDriver(r.adaptor, relay.Gpio)
		relay.Start()
		relay.Off()
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-c:
		failSafe()
	}
}
