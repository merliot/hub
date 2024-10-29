package hub

import (
	"html/template"
	"math"
	"net/url"
	"sync"
	"time"
)

var (
	runningSite bool // running as full web-site
	runningDemo bool // running in DEMO mode
)

// Devicer is the device model interface.  A device is a concrete Devicer.
type Devicer interface {
	// GetConfig returns the device configuration
	GetConfig() Config
	// GetHandlers returns the device message Handlers
	GetHandlers() Handlers
	// Setup prepares the device for operation.  Device hardware and other
	// initializations are done here.  Returning an error fails the device load.
	Setup() error
	// Poll services the device.  Poll is called every Config.PollPeriod
	// seconds.  The Packet can be used to send a message.
	Poll(*Packet)
	// DemoSetup is DEMO mode Setup
	DemoSetup() error
	// DemoPoll is DEMO mode Poll
	DemoPoll(*Packet)
}

type device struct {
	Id           string
	Model        string
	Name         string
	Children     []string
	DeployParams template.HTML
	flags        `json:"-"`
	Config       `json:"-"`
	Devicer      `json:"-"`
	Handlers     `json:"-"`
	sync.RWMutex `json:"-"`
	deviceOS
	stopChan chan struct{}
	startup  time.Time
}

func (d *device) build(maker Maker) error {

	d.startup = time.Now()
	d.Devicer = maker()
	d.Config = d.GetConfig()
	d.Handlers = d.GetHandlers()
	d.flags = d.Config.Flags

	if runningSite {
		d.Set(flagLocked)
	}

	if runningDemo {
		d.Set(flagDemo | flagOnline | flagMetal)
	}

	// Special handlers
	d.Handlers["/state"] = &Handler[any]{d.handleState}
	d.Handlers["/reboot"] = &Handler[NoMsg]{d.handleReboot}
	d.Handlers["/get-uptime"] = &Handler[NoMsg]{d.handleGetUptime}
	d.Handlers["/uptime"] = &Handler[msgUptime]{d.handleUptime}

	// Bracket poll period: [1..forever) seconds
	if d.PollPeriod == 0 {
		d.PollPeriod = time.Duration(math.MaxInt64)
	} else if d.PollPeriod < time.Second {
		d.PollPeriod = time.Second
	}

	// Configure the device using DeployParams
	_, err := d._formConfig(string(d.DeployParams))
	if err != nil {
		LogError("Configuring device using DeployParams", "err", err, "device", d)
		// TODO should there be a return here?
	}

	return d.buildOS()
}

func (d *device) handleState(pkt *Packet) {
	pkt.Unmarshal(d.State).RouteUp()
}

type msgUptime struct {
	time.Duration
}

func (d *device) handleGetUptime(pkt *Packet) {
	var uptime = time.Since(d.startup)
	var msg = msgUptime{uptime}
	pkt.SetPath("/uptime").Marshal(&msg).RouteUp()
}

func (d *device) handleUptime(pkt *Packet) {
	var msg msgUptime
	pkt.Unmarshal(&msg)
	d.startup = time.Now().Add(-msg.Duration)
	pkt.RouteUp()
}

func (d *device) _formConfig(rawQuery string) (changed bool, err error) {

	// rawQuery is the proposed new DeployParams
	proposedParams, err := url.QueryUnescape(rawQuery)
	if err != nil {
		return false, err
	}
	values, err := url.ParseQuery(proposedParams)
	if err != nil {
		return false, err
	}

	//	LogInfo("Proposed DeployParams:", proposedParams)

	// Form-decode these values into the device to configure the device
	if err := decoder.Decode(d.State, values); err != nil {
		return false, err
	}

	if proposedParams == string(d.DeployParams) {
		// No change
		return false, nil
	}

	// Save changes.  Store DeployParams unescaped.
	d.DeployParams = template.HTML(proposedParams)
	return true, nil
}

func (d *device) formConfig(rawQuery string) (changed bool, err error) {
	d.Lock()
	defer d.Unlock()
	return d._formConfig(rawQuery)
}
