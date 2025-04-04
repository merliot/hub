//go:build !tinygo

package device

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
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
	//d.server.logDebug("Handle", "pattern", pattern)
	d.ServeMux.Handle(pattern, handler)
}

func (d *device) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	//d.server.logDebug("HandleFunc", "pattern", pattern)
	d.ServeMux.HandleFunc(pattern, handler)
}

func (s *server) buildOS(d *device) error {
	var err error

	d.ServeMux = http.NewServeMux()

	// Build device's layered FS.  fs is stacked on top of
	// deviceFs, so fs:foo.tmpl will override deviceFs:foo.tmpl,
	// when searching for file foo.tmpl.
	d.layeredFS.stack(deviceFs)
	if d.FS != nil {
		d.layeredFS.stack(d.FS)
	}

	// Merge device-specific funcs with base server and base device funcs
	// to make one FuncMap
	if d.FuncMap == nil {
		d.FuncMap = make(FuncMap)
	}
	for k, v := range s.baseFuncs() {
		d.FuncMap[k] = v
	}
	for k, v := range d.baseFuncs() {
		d.FuncMap[k] = v
	}

	// Build the device templates
	d.templates, err = d.layeredFS.parseFS("template/*.tmpl",
		template.FuncMap(d.FuncMap))

	return err
}

func (s *server) addChild(parent *device, id, model, name string, flags flags) (err error) {

	var resurrect bool

	if _, exists := parent.children.get(id); exists {
		return fmt.Errorf("Device's children already includes child id %s", id)
	}

	child, exists := s.devices.get(id)
	if exists {
		if !child.isSet(flagGhost) {
			return fmt.Errorf("Child device id %s already exists", id)
		}
		// Child exists, but it's a ghost: resurrect
		resurrect = true
	} else {
		child, err = s.newDevice(id, model, name)
		if err != nil {
			return
		}
	}

	if resurrect {
		// Resurrect ghost child back to life
		child.unSet(flagGhost)
	}

	m, exists := s.models.get(model)
	if !exists {
		return fmt.Errorf("Model '%s' not registered", model)
	}

	child.model = m
	if err = s.build(child, s.defaultDeviceFlags()); err != nil {
		return
	}

	child.set(flags)
	child.parent = parent
	parent.Children = append(parent.Children, id)
	parent.children.Store(id, child)
	s.devices.Store(id, child)

	child.installAPI()

	if !resurrect {
		// Only install /device/{id} pattern if not previously ghosted
		s.deviceInstall(child)
	}

	if s.isSet(flagRunningDemo) {
		if err = child.demoSetup(); err != nil {
			return
		}
		child.startDemo()
	}

	return
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

	if s.isSet(flagRunningDemo) {
		child.stop()
	}

	// Remove child from parent
	child.orphan()

	// We don't remove the child completely because /device/{childId} is
	// already installed to point to this child.  So mark the child (and all
	// of its descendents) as ghosts.  Later, if the child is added back,
	// we'll resurrect it.
	child.ghost()
}

func (d *device) _familyTree(devs devicesJSON) {
	if !d.isSet(flagGhost) {
		devs[d.Id] = d
		d.children.drange(func(_ string, child *device) bool {
			child._familyTree(devs)
			return true
		})
	}
}

func (d *device) familyTree() devicesJSON {
	devs := make(devicesJSON)
	d._familyTree(devs)
	return devs
}

func (s *server) deviceOffline(id string) {
	if d, exists := s.devices.get(id); exists {
		d.unSet(flagOnline)
		pkt := d.newPacket()
		pkt.SetPath("/offline").BroadcastUp()
	}
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
	var noEnv bool = (s.devicesEnv == "")
	var noFile bool = (s.devicesFile == "")
	var noDefault bool

	defaultFile, err := os.Open("devices.json")
	if err == nil {
		defaultFile.Close()
	}
	noDefault = (err != nil)

	switch {

	case noEnv && noFile && noDefault:
		s.logInfo("Loading with empty hub")
		if !s.isSet(flagAutoSave) {
			s.set(flagSaveToClipboard)
		}
		if err := json.Unmarshal([]byte(emptyHub), &devs); err != nil {
			return err
		}

	case noEnv && noFile && !noDefault:
		s.logInfo("Loading from devices.json")
		if err := fileReadJSON("devices.json", &devs); err != nil {
			return err
		}

	case noEnv:
		s.logInfo("Loading from", "DEVICES_FILE", s.devicesFile)
		if err := fileReadJSON(s.devicesFile, &devs); err != nil {
			return err
		}

	default:
		s.logInfo("Loading from DEVICES env var")
		s.set(flagSaveToClipboard)
		if err := json.Unmarshal([]byte(s.devicesEnv), &devs); err != nil {
			return err
		}
	}

	s.devices.loadJSON(devs)
	return nil
}

func (s *server) dirty() error {
	s.set(flagDirty)
	pkt := s.newPacket().SetDst(s.root.Id).SetPath("/dirty")
	return s.sessions.routeAll(pkt)
}

func (s *server) clean() error {
	s.unSet(flagDirty)
	pkt := s.newPacket().SetDst(s.root.Id).SetPath("/clean")
	return s.sessions.routeAll(pkt)
}

func (s *server) devicesSave() error {
	var noEnv bool = (s.devicesEnv == "")
	var noFile bool = (s.devicesFile == "")

	if noEnv && noFile {
		//s.logDebug("Saving to devices.json")
		if err := fileWriteJSON("devices.json", s.devices.getJSON()); err != nil {
			return err
		}
		return s.clean()
	}

	if noEnv && !noFile {
		//s.logDebug("Saving to", "DEVICES_FILE", s.devicesFile)
		if err := fileWriteJSON(s.devicesFile, s.devices.getJSON()); err != nil {
			return err
		}
		return s.clean()
	}

	// Save to clipboard

	return nil
}

func (d *device) demoReboot(pkt *Packet) {
	// Simulate a reboot
	d.stop()

	// Go offline for 3 seconds
	d.unSet(flagOnline)
	pkt.SetPath("/offline").BroadcastUp()
	time.Sleep(3 * time.Second)

	d.startup = time.Now()
	d.formConfig(string(d.DeployParams))

	d.DemoSetup()
	d.startDemo()

	// Come back online
	d.set(flagOnline)
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
