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

func (d *device) buildOS() error {
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

	return err
}

func (s *server) addChild(parent *device, id, model, name string, flags flags) error {

	var resurrect bool

	maker, ok := s.models[model]
	if !ok {
		return fmt.Errorf("Unknown model")
	}

	if _, exists := parent.children.load(id); exists {
		return fmt.Errorf("Device's children already includes child")
	}

	child, exists := s.devices.load(id)
	if exists {
		if !child.isSet(flagGhost) {
			return fmt.Errorf("Child device already exists")
		}
		// Child exists, but it's a ghost: resurrect
		resurrect = true
	} else {
		child = &device{Id: id, Model: model, Name: name}
	}

	if resurrect {
		// Resurrect ghost child back to life
		child.unSet(flagGhost)
	}

	if err := child.build(maker.Maker); err != nil {
		return err
	}

	child.set(flags)
	child.parent = parent
	parent.Children = append(parent.Children, id)
	parent.children.Store(id, child)
	s.devices.Store(id, child)

	child.setupAPI()

	if !resurrect {
		// Only install /device/{id} pattern if not previously ghosted
		child.install()
	}

	if s.runningDemo {
		if err := child.setup(); err != nil {
			return err
		}
		child.startDemo()
	}

	return nil
}

// ghost child and all children, recursively
func (child *device) ghost() {
	child.children.drange(func(id string, d *device) bool {
		d.ghost()
		return true
	})
	child.set(flagGhost)
	child.parent = nil
	child.nexthop = nil
	child.DeployParams = ""
	child.Children = []string{}
	child.children.Clear()
}

// orphan detaches child from its parent
func (child *device) orphan() {
	parent := child.parent
	if parent != nil {
		parent.children.Delete(child.Id)
		if index := slices.Index(parent.Children, child.Id); index != -1 {
			parent.Children = slices.Delete(parent.Children, index, index+1)
		}
	}
}

func (s *server) removeChild(child *device) {

	if s.runningDemo {
		child.stopDemo()
	}

	// Remove child from parent
	child.orphan()

	// We don't remove the child completely because /device/{childId} is
	// already installed to point to this child.  So mark the child (and all
	// of its descendents) as ghosts.  Later, if the child is added back,
	// we'll resurrect it.
	child.ghost()
}

func deviceRenderPkt(w io.Writer, sessionId string, pkt *Packet) error {
	//LogDebug("deviceRenderPkt", "sessionId", sessionId, "pkt", pkt)
	d, err := getDevice(pkt.Dst)
	if err != nil {
		return err
	}
	return d.renderPkt(w, sessionId, pkt)
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

	s.devices.loadJSON(devs)
	return nil
}

func (s *server) devicesSave() error {
	var devicesJSON = Getenv("DEVICES", "")
	var devicesFile = Getenv("DEVICES_FILE", "")
	var noJSON bool = (devicesJSON == "")
	var noFile bool = (devicesFile == "")

	if noJSON && noFile {
		//LogDebug("Saving to devices.json")
		return fileWriteJSON("devices.json", s.devices.getJSON())
	}

	if noJSON && !noFile {
		//LogDebug("Saving to", "DEVICES_FILE", devicesFile)
		return fileWriteJSON(devicesFile, s.devices.getJSON())
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
