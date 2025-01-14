package device

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"time"
)

var (
	runningSite bool // running as full web-site
	runningDemo bool // running in DEMO mode
)

// Devicer is the device model interface
type Devicer interface {
	// GetConfig returns the device configuration
	GetConfig() Config
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
	Config       `json:"-"`
	Devicer      `json:"-"`
	stopChan     chan struct{}
	startup      time.Time
	stateMu      mutex
	flags
	rwMutex
	deviceOS
}

func (d *device) String() string {
	return fmt.Sprintf("[%s:%s:%s]", d.Id, d.Model, d.Name)
}

func (d *device) _build(maker Maker) error {

	d.startup = time.Now()
	d.Devicer = maker()
	d.Config = d.GetConfig()
	d.flags = d.Config.Flags

	if d.PacketHandlers == nil {
		d.PacketHandlers = PacketHandlers{}
	}

	if runningSite {
		d._set(flagLocked)
	}
	if runningDemo {
		d._set(flagDemo | flagOnline | flagMetal)
	}

	// Special handlers
	d.PacketHandlers["/state"] = &PacketHandler[any]{d.handleState}
	d.PacketHandlers["/online"] = &PacketHandler[any]{d.handleOnline}
	d.PacketHandlers["/offline"] = &PacketHandler[any]{d.handleOffline}
	d.PacketHandlers["/reboot"] = &PacketHandler[NoMsg]{d.handleReboot}
	d.PacketHandlers["/get-uptime"] = &PacketHandler[NoMsg]{d.handleGetUptime}
	d.PacketHandlers["/uptime"] = &PacketHandler[msgUptime]{d.handleUptime}

	// Bracket poll period: [1..forever) seconds
	if d.PollPeriod == 0 {
		d.PollPeriod = time.Duration(math.MaxInt64)
	} else if d.PollPeriod < time.Second {
		d.PollPeriod = time.Second
	}

	// Configure the device using DeployParams
	_, err := d._formConfig(string(d.DeployParams))
	if err != nil {
		LogError("Configuring device using DeployParams",
			"device", d, "err", err)
	}

	return d._buildOS()
}

func (d *device) build(maker Maker) error {
	d.Lock()
	defer d.Unlock()
	return d._build(maker)
}

func (d *device) handleState(pkt *Packet) {
	pkt.Unmarshal(d.State).BroadcastUp()
}

func (d *device) handleOnline(pkt *Packet) {
	d._set(flagOnline)
	pkt.Unmarshal(d.State).BroadcastUp()
}

func (d *device) handleOffline(pkt *Packet) {
	d._unSet(flagOnline)
	pkt.BroadcastUp()
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

	//LogError("Proposed", "DeployParams", proposedParams, "values", values)

	// Form-decode these values into the device to configure the device
	if err := decode(d.State, values); err != nil {
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
