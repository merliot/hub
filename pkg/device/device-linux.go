//go:build !tinygo

package device

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"sync"
	"time"
)

//go:embed robots.txt blog css docs images js template
var deviceFs embed.FS

type deviceOS struct {
	*http.ServeMux
	templates *template.Template
	layeredFS
	views sync.Map
}

func (d *device) Handle(pattern string, handler http.Handler) {
	//LogDebug("Handle", "pattern", pattern)
	d.ServeMux.Handle(pattern, handler)
}

func (d *device) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	//LogDebug("HandleFunc", "pattern", pattern)
	d.ServeMux.HandleFunc(pattern, handler)
}

func (d *device) _buildOS() error {
	var err error

	d.ServeMux = http.NewServeMux()

	// Build device's layered FS.  fs is stacked on top of
	// deviceFs, so fs:foo.tmpl will override deviceFs:foo.tmpl,
	// when searching for file foo.tmpl.
	d.layeredFS.stack(deviceFs)
	if d.FS != nil {
		d.layeredFS.stack(d.FS)
	}

	// Merge base funcs with device funcs to make one FuncMap
	if d.FuncMap == nil {
		d.FuncMap = template.FuncMap{}
	}
	for k, v := range d.baseFuncs() {
		d.FuncMap[k] = v
	}

	// Build the device templates using combined funcs
	d.templates, err = d.layeredFS.parseFS("template/*.tmpl", d.FuncMap)

	// Default handlers for Linux devices
	d.PacketHandlers["/created"] = &PacketHandler[MsgCreated]{d.handleCreated}
	d.PacketHandlers["/destroyed"] = &PacketHandler[MsgDestroyed]{d.handleDestroyed}
	d.PacketHandlers["/downloaded"] = &PacketHandler[MsgDownloaded]{d.handleDownloaded}
	d.PacketHandlers["/announced"] = &PacketHandler[MsgDownloaded]{d.handleAnnounced}

	return err
}

func (d *device) _devices(family deviceMap) {
	d.RLock()
	defer d.RUnlock()

	family[d.Id] = devices[d.Id]
	for _, childId := range d.Children {
		devices[childId]._devices(family)
	}
}

// devices returns a deviceMap with this device and all its children
func (d *device) devices() deviceMap {
	var family = make(deviceMap)

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	d._devices(family)

	return family
}

func (d *device) _setupAPI() {
	// Install base + device APIs
	d.installAPIs()
	// Install the device-specific packet handlers APIs
	d.packetHandlersInstall()
}

func (d *device) setupAPI() {
	d.Lock()
	d._setupAPI()
	d.Unlock()
}

func addChild(parent *device, id, model, name string, flags flags) error {

	var resurrect bool

	maker, ok := Models[model]
	if !ok {
		return fmt.Errorf("Unknown model")
	}

	parent.Lock()
	defer parent.Unlock()

	if slices.Contains(parent.Children, id) {
		return fmt.Errorf("Device's children already includes child")
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()

	child, ok := devices[id]
	if ok {
		if !child.isSet(flagGhost) {
			return fmt.Errorf("Child device already exists")
		}
		// Child exists, but it's a ghost: resurrect
		resurrect = true
	} else {
		child = &device{Id: id, Model: model, Name: name}
	}

	child.Lock()
	defer child.Unlock()

	if resurrect {
		// Resurrect ghost child back to life
		child._unSet(flagGhost)
		child.DeployParams = ""
		child.Children = []string{}
	}

	if err := child._build(maker.Maker); err != nil {
		return err
	}

	child._set(flags)

	parent.Children = append(parent.Children, id)
	devices[id] = child

	child._setupAPI()

	if !resurrect {
		child.deviceInstall()
	}

	if runningDemo {
		if err := child._setup(); err != nil {
			return err
		}
		child._startDemo()
	}

	return nil
}

func removeChild(id string) error {

	// Ghost all children
	if err := ghostChildren(id); err != nil {
		return err
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()

	d := devices[id]

	if runningDemo {
		d.stopDemo()
	}

	// Ghost device
	d.set(flagGhost)

	// Detach the device from any parent devices where it is listed as a child
	for _, d := range devices {
		d.Lock()
		if index := slices.Index(d.Children, id); index != -1 {
			d.Children = slices.Delete(d.Children, index, index+1)
		}
		d.Unlock()
	}

	return nil
}

func (d *device) routeDown(pkt *Packet) {

	// If device is running on 'metal', this is the packet's final
	// destination.
	if d.isSet(flagMetal) {
		d.handle(pkt)
		return
	}

	// Otherwise, route the packet down
	downlinkRoute(d.Id, pkt)
}

func deviceRouteDown(id string, pkt *Packet) {
	if d, err := getDevice(id); err == nil {
		d.routeDown(pkt)
	}
}

func deviceRouteUp(id string, pkt *Packet) {
	if d, err := getDevice(id); err == nil {
		d.handle(pkt)
	}
}

func deviceRenderPkt(w io.Writer, sessionId string, pkt *Packet) error {
	//LogDebug("deviceRenderPkt", "sessionId", sessionId, "pkt", pkt)
	d, err := getDevice(pkt.Dst)
	if err != nil {
		return err
	}
	return d.renderPkt(w, sessionId, pkt)
}

func _ghostChild(id string) error {

	d, ok := devices[id]
	if !ok {
		return deviceNotFound(id)
	}

	if runningDemo {
		d.stopDemo()
	}

	d.Lock()
	defer d.Unlock()

	d._set(flagGhost)

	for _, childId := range d.Children {
		if err := _ghostChild(childId); err != nil {
			return err
		}
	}

	d.Children = []string{}

	return nil
}

func _ghostChildren(id string) error {

	d, ok := devices[id]
	if !ok {
		return deviceNotFound(id)
	}

	d.Lock()
	defer d.Unlock()

	for _, childId := range d.Children {
		if err := _ghostChild(childId); err != nil {
			return err
		}
	}

	d.Children = []string{}

	return nil
}

func ghostChildren(id string) error {
	devicesMu.Lock()
	defer devicesMu.Unlock()
	return _ghostChildren(id)
}

func (d *device) _copyDevice(from *device) {
	d.Model = from.Model
	d.Name = from.Name
	d.Children = from.Children
	d.DeployParams = from.DeployParams
	d.Config = from.Config
	d.flags = from.flags
}

func merge(devices, newDevices deviceMap) error {

	// Find the root (anchor) of the new device tree
	anchor, err := findRoot(newDevices)
	if err != nil {
		return err
	}

	// Recursively ghost the children of the anchor device in the existing
	// device tree.  The children may be resurrected while merging if they
	// exists in the new device tree.
	if err := ghostChildren(anchor.Id); err != nil {
		return err
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()

	// Swing anchor to existing tree
	anchor = devices[anchor.Id]

	anchor.Lock()
	defer anchor.Unlock()

	// Now merge in the new devices, setting up each device as we go
	for newId, newDevice := range newDevices {

		if newId == anchor.Id {
			anchor.Children = newDevice.Children
			// All we want for the anchor is the new anchor child list
			continue
		}

		device, exists := devices[newId]
		if exists {
			// Better be a ghost
			if !device.isSet(flagGhost) {
				return fmt.Errorf("Device %s already exists, aborting merge", device)
			}
		} else {
			device = newDevice
		}

		device.Lock()

		device._copyDevice(newDevice)

		maker, ok := Models[device.Model]
		if !ok {
			device.Unlock()
			return fmt.Errorf("Unknown model")
		}

		if err := device._build(maker.Maker); err != nil {
			device.Unlock()
			return err
		}

		device._set(flagLocked)

		devices[newId] = device

		device._setupAPI()

		if !exists {
			device.deviceInstall()
		}

		if runningDemo {
			if err := device._setup(); err != nil {
				device.Unlock()
				return err
			}
			device._startDemo()
		}

		device.Unlock()
	}

	anchor._set(flagOnline)

	return nil
}

func validate(a *device) error {

	d, err := getDevice(a.Id)
	if err != nil {
		return err
	}

	d.RLock()
	defer d.RUnlock()

	if d.Model != a.Model {
		return fmt.Errorf("Device model wrong.  Want %s; got %s",
			d.Model, a.Model)
	}

	if d.Name != a.Name {
		return fmt.Errorf("Device name wrong.  Want %s; got %s",
			d.Name, a.Name)
	}

	if d.DeployParams != a.DeployParams {
		return fmt.Errorf("Device DeployParams wrong.\nWant: %s\nGot: %s",
			d.DeployParams, a.DeployParams)
	}

	return nil
}

func deviceOffline(id string) {
	if d, err := getDevice(id); err == nil {
		d.unSet(flagOnline)
		pkt := &Packet{Dst: id, Path: "/offline"}
		pkt.BroadcastUp()
	}
}

func (d *device) dirty() {
	d.set(flagDirty)
	pkt := &Packet{Dst: d.Id, Path: "/dirty"}
	pkt.BroadcastUp()
}

func (d *device) clean() {
	d.unSet(flagDirty)
	pkt := &Packet{Dst: d.Id, Path: "/dirty"}
	pkt.BroadcastUp()
}

func deviceParent(id string) string {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	for _, d := range devices {
		d.RLock()
		if slices.Contains(d.Children, id) {
			d.RUnlock()
			return d.Id
		}
		d.RUnlock()
	}
	return ""
}

var emptyHub = `{
	"hub": {
		"Id": "hub",
		"Model": "hub",
		"Name": "Hub",
		"Children": [],
		"DeployParams": "target=x86-64&port=8000"
	}
}`

func (s *server) loadDevices() error {

	var devs = make(devicesJSON)
	var autoSave = Getenv("AUTO_SAVE", "true") == "true"
	var devicesEnv = Getenv("DEVICES", "")
	var devicesFile = Getenv("DEVICES_FILE", "")
	var noEnv bool = (devicesEnv == "")
	var noFile bool = (devicesFile == "")
	var noDefault bool

	defaultFile, err := os.Open("devices.json")
	if err == nil {
		defaultFile.Close()
	}
	noDefault = (err != nil)

	switch {

	case noEnv && noFile && noDefault:
		LogInfo("Loading with empty hub")
		s.saveToClipboard = !autoSave
		if err := json.Unmarshal([]byte(emptyHub), &devs); err != nil {
			return err
		}

	case noEnv && noFile && !noDefault:
		LogInfo("Loading from devices.json")
		if err := fileReadJSON("devices.json", &devs); err != nil {
			return err
		}

	case noEnv:
		LogInfo("Loading from", "DEVICES_FILE", devicesFile)
		if err := fileReadJSON(devicesFile, &devs); err != nil {
			return err
		}

	default:
		LogInfo("Loading from DEVICES env var")
		s.saveToClipboard = true
		if err := json.Unmarshal([]byte(devicesEnv), &devs); err != nil {
			return err
		}
	}

	s.storeJSON(devs)
	return nil
}

func (d *device) save() error {
	var autoSave = Getenv("AUTO_SAVE", "true") == "true"

	if autoSave {
		return devicesSave()
	}

	// Mark device dirty so user can manually save
	d.dirty()

	return nil
}

func (s *server) devicesSave() error {
	var devicesJSON = Getenv("DEVICES", "")
	var devicesFile = Getenv("DEVICES_FILE", "")
	var noJSON bool = (devicesJSON == "")
	var noFile bool = (devicesFile == "")

	if noJSON && noFile {
		//LogDebug("Saving to devices.json")
		return fileWriteJSON("devices.json", s.aliveDevices())
	}

	if noJSON && !noFile {
		//LogDebug("Saving to", "DEVICES_FILE", devicesFile)
		return fileWriteJSON(devicesFile, s.aliveDevices())
	}

	// Save to clipboard

	return nil
}

func (s *server) devicesOnline(l linker) {

	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()

	for id, d := range s.devices {
		var pkt = &Packet{
			Dst:  id,
			Path: "/online",
		}
		d.RLock()
		if !d._isSet(flagOnline) {
			d.RUnlock()
			continue
		}
		pkt.Marshal(d.State)
		d.RUnlock()
		LogInfo("Sending", "pkt", pkt)
		l.Send(pkt)
	}
}

func (d *device) demoReboot(pkt *Packet) {
	// Simulate a reboot
	d.stopDemo()

	// Go offline for 3 seconds
	d.unSet(flagOnline)
	pkt.SetPath("/offline").BroadcastUp()
	time.Sleep(3 * time.Second)

	model, _ := Models[d.Model]

	d.build(model.Maker)
	d.setupAPI()
	d.setup()
	d.startDemo()

	// Come back online
	pkt.SetPath("/online").Marshal(d.State).BroadcastUp()
}

func (d *device) handleReboot(pkt *Packet) {
	if d.isSet(flagDemo) {
		d.demoReboot(pkt)
	} else {
		// Exit the app, stopping the device.  A systemd script will
		// restart the device.
		os.Exit(0)
	}
}

func devicesSortedId() []string {
	keys := make([]string, 0, len(devices))
	for id := range devices {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	return keys
}

type deviceStatus struct {
	Color  string
	Status string
}

func devicesStatus() []deviceStatus {
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	var statuses = make([]deviceStatus, len(devices))
	for _, id := range devicesSortedId() {
		d := devices[id]
		d.RLock()
		status := fmt.Sprintf("%-16s %-16s %-16s %3d",
			d.Id, d.Model, d.Name, len(d.Children))
		color := "gold"
		if d._isSet(flagGhost) {
			color = "gray"
		}
		statuses = append(statuses, deviceStatus{
			Color:  color,
			Status: status,
		})
		d.RUnlock()
	}
	return statuses
}

/*
func dumpStack() {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	println("Stack:\n%s", string(buf))
}
*/
