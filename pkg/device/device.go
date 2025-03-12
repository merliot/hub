package device

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"time"
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
	DeployParams template.URL
	Config       `json:"-"`
	Devicer      `json:"-"`
	model        Model
	children     deviceMap
	parent       *device
	nexthop      *device
	stopChan     chan struct{}
	startup      time.Time
	stateMu      mutex
	flags
	deviceOS
}

func (d *device) String() string {
	return fmt.Sprintf("[%s:%s:%s]", d.Id, d.Model, d.Name)
}

func (d *device) build(model Model, additionalFlags flags) error {

	d.startup = time.Now()
	d.model = model
	d.Devicer = model.Maker()
	d.Config = d.GetConfig()
	d.flags = d.Config.Flags
	d.set(additionalFlags)

	if d.APIs == nil {
		d.APIs = APIs{}
	}

	if d.PacketHandlers == nil {
		d.PacketHandlers = PacketHandlers{}
	}

	// Default handlers for all devices
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
	_, err := d.formConfig(string(d.DeployParams))
	if err != nil {
		LogError("Configuring device using DeployParams",
			"device", d, "err", err)
	}

	return d.buildOS()
}

func (d *device) handleState(pkt *Packet) {
	pkt.Unmarshal(d.State).BroadcastUp()
}

func (d *device) handleOnline(pkt *Packet) {
	d.set(flagOnline)
	pkt.Unmarshal(d.State).BroadcastUp()
}

func (d *device) handleOffline(pkt *Packet) {
	d.unSet(flagOnline)
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

func (d *device) formConfig(rawQuery string) (changed bool, err error) {

	// rawQuery is the proposed new DeployParams
	proposedParams, err := url.QueryUnescape(rawQuery)
	if err != nil {
		return false, err
	}
	values, err := url.ParseQuery(proposedParams)
	if err != nil {
		return false, err
	}

	//LogDebug("Proposed", "DeployParams", proposedParams, "values", values)

	// Form-decode these values into the device to configure the device
	if err := decode(d.State, values); err != nil {
		return false, err
	}

	/*
		// Form-encode the device to the full values
		values, err = encode(d.State)
		if err != nil {
			return false, err
		}

		proposedParams = values.Encode()

		LogDebug("Proposed", "DeployParams", proposedParams, "values", values)
	*/

	if template.URL(proposedParams) == d.DeployParams {
		// No change
		return false, nil
	}

	// Save changes.  Store DeployParams unescaped.
	d.DeployParams = template.URL(proposedParams)
	return true, nil
}

func validateIds(id, name string) error {
	if err := validateId(id); err != nil {
		return err
	}
	if err := validateName(name); err != nil {
		return err
	}
	return nil
}
