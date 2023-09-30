//go:build !tinygo

package relays

import (
	"embed"
	"html/template"
	"net/http"
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

type relaysOS struct {
	templates *template.Template
	adaptor   *raspi.Adaptor
}

func (r *Relays) relaysOSNew() {
	r.CompositeFs.AddFS(fs)
	r.templates = r.CompositeFs.ParseFS("template/*")
	r.adaptor = raspi.NewAdaptor()
}

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

func (r *Relays) pins() common.GpioPins {
	return r.Targets["rpi"].GpioPins
}

// FailSafe by turning off all gpios
func (r *Relays) failSafe() {
	for _, pin := range r.pins() {
		rpin := strconv.Itoa(pin)
		driver := gpio.NewRelayDriver(r.adaptor, rpin)
		driver.Start()
		driver.Off()
	}
}

func (r *Relays) runOS(i *dean.Injector) {

	defer func() {
		if recover() != nil {
			r.failSafe()
		}
	}()

	r.adaptor.Connect()

	for i := range r.Relays {
		relay := &r.Relays[i]
		if relay.Gpio == "" {
			continue
		}
		if pin, ok := r.pins()[relay.Gpio]; ok {
			rpin := strconv.Itoa(pin)
			relay.driver = gpio.NewRelayDriver(r.adaptor, rpin)
			relay.Start()
			relay.Off()
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-c:
		r.failSafe()
	}
}
