//go:build !tinygo

package device

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/merliot/hub/pkg/ratelimit"
)

// Server flags
const (
	flagRunningDemo     flags = 1 << iota // Running in DEMO mode
	flagRunningSite                       // Running in SITE mode
	flagSaveToClipboard                   // Save changes to clipboard
	flagAutoSave                          // Automatically save changes
	flagDirty                             // Server has unsaved changes
	flagDebugKeepBuilds                   // Don't delete temp build directory
)

type server struct {
	devices        deviceMap      // key: device id, value: *device
	sessions       sessionMap     // key: session id, value: *session
	uplinks        uplinkMap      // key: linker
	downlinks      downlinkMap    // key: device id, value: linker
	models         modelMap       // key: model name, value: *model
	packetHandlers PacketHandlers // key: path, value: handler
	root           *device
	mux            *http.ServeMux
	server         *http.Server
	flags
	port            int
	wsxPingPeriod   int
	background      string
	wifiSsids       []string
	wifiPassphrases []string
	devicesEnv      string
	devicesFile     string
	logLevel        string
	dialUrls        string
	user            string
	passwd          string
}

type ServerOption func(*server)

func WithPort(port int) ServerOption {
	return func(s *server) {
		s.port = port
	}
}

func WithModels(models Models) ServerOption {
	return func(s *server) {
		s.models.load(models)
	}
}

func WithPingPeriod(period int) ServerOption {
	return func(s *server) {
		if period < 2 {
			period = 2
		}
		s.wsxPingPeriod = period
	}
}

func WithKeepBuilds(keep string) ServerOption {
	return func(s *server) {
		if keep == "true" {
			s.set(flagDebugKeepBuilds)
		}
	}
}

func WithRunningSite(running string) ServerOption {
	return func(s *server) {
		if running == "true" {
			s.set(flagRunningSite)
		}
	}
}

func WithRunningDemo(running string) ServerOption {
	return func(s *server) {
		if running == "true" {
			s.set(flagRunningDemo)
		}
	}
}

func WithBackground(bg string) ServerOption {
	return func(s *server) {
		s.background = bg
	}
}

func WithWifiSsids(ssids string) ServerOption {
	return func(s *server) {
		if ssids != "" {
			s.wifiSsids = strings.Split(ssids, ",")
		}
	}
}

func WithWifiPassphrases(ps string) ServerOption {
	return func(s *server) {
		if ps != "" {
			s.wifiPassphrases = strings.Split(ps, ",")
		}
	}
}

func WithAutoSave(save string) ServerOption {
	return func(s *server) {
		if save == "true" {
			s.set(flagAutoSave)
		}
	}
}

func WithDevicesEnv(devs string) ServerOption {
	return func(s *server) {
		s.devicesEnv = devs
	}
}

func WithDevicesFile(file string) ServerOption {
	return func(s *server) {
		s.devicesFile = file
	}
}

func WithLogLevel(level string) ServerOption {
	return func(s *server) {
		s.logLevel = level
	}
}

func WithDialUrls(urls string) ServerOption {
	return func(s *server) {
		s.dialUrls = urls
	}
}

func WithUser(user string) ServerOption {
	return func(s *server) {
		s.user = user
	}
}

func WithPasswd(passwd string) ServerOption {
	return func(s *server) {
		s.passwd = passwd
	}
}

var rlConfig = ratelimit.Config{
	FillInterval:    100 * time.Millisecond,
	Capacity:        30,
	CleanupInterval: 1 * time.Minute,
}

// NewServer returns a device server listening on addr
func NewServer(options ...ServerOption) *server {

	s := &server{
		packetHandlers: make(PacketHandlers),
		mux:            http.NewServeMux(),
		server:         &http.Server{},
		wsxPingPeriod:  30,
		logLevel:       "INFO",
	}

	for _, opt := range options {
		opt(s)
	}

	if s.isSet(flagRunningSite) {
		s.set(flagRunningDemo)
	}

	s.packetHandlers["/created"] = &PacketHandler[msgCreated]{s.handleCreated}
	s.packetHandlers["/destroyed"] = &PacketHandler[msgDestroy]{s.handleDestroyed}
	s.packetHandlers["/downloaded"] = &PacketHandler[msgDownloaded]{s.handleDownloaded}
	s.packetHandlers["/announced"] = &PacketHandler[deviceMap]{s.handleAnnounced}

	rl := ratelimit.New(rlConfig)
	s.server.Handler = rl.RateLimit(s.basicAuth(s.mux))

	s.sessions.start()

	return s
}

// routeDown routes the packet down to a downlink.  Which downlink is
// determined by a lookup in the routing table for the "next-hop" downlink, the
// downlink which is towards the destination.
func (s *server) routeDown(pkt *Packet) error {
	s.logDebug("routeDown", "pkt", pkt)

	d, ok := s.devices.get(pkt.Dst)
	if !ok {
		return deviceNotFound(pkt.Dst)
	}

	nexthop := d.nexthop
	if nexthop.isSet(flagMetal) {
		nexthop.handle(pkt)
	} else {
		s.downlinks.route(nexthop.Id, pkt)
	}

	return nil
}

func (s *server) defaultDeviceFlags() flags {
	var flags flags
	if s.isSet(flagRunningSite) {
		flags = flagLocked
	}
	if s.isSet(flagRunningDemo) {
		flags = flagDemo | flagOnline | flagMetal
	}
	return flags
}

func (s *server) buildDevice(id string, d *device) error {
	if id != d.Id {
		return fmt.Errorf("Mismatching Ids")
	}
	if err := validateId(id); err != nil {
		return err
	}
	if err := validateName(d.Name); err != nil {
		return err
	}
	model, exists := s.models.get(d.Model)
	if !exists {
		return fmt.Errorf("Model '%s' not registered", d.Model)
	}
	d.model = model
	d.server = s
	return s.build(d, s.defaultDeviceFlags())
}

func (s *server) newDevice(id, model, name string) (d *device, err error) {
	if err = validateId(id); err != nil {
		return
	}
	if err = validateName(name); err != nil {
		return
	}
	m, exists := s.models.get(model)
	if !exists {
		return nil, fmt.Errorf("Model '%s' not registered", model)
	}

	d = &device{
		Id:     id,
		Model:  model,
		Name:   name,
		model:  m,
		server: s,
	}

	return
}

func (s *server) buildDevices() {
	s.devices.drange(func(id string, d *device) bool {
		if err := s.buildDevice(id, d); err != nil {
			s.logError("Skipping", "device", d, "err", err)
			s.devices.Delete(id)
		}
		return true
	})
}

func (s *server) buildTree() (err error) {
	s.root, err = s.devices.buildTree()
	return
}

// Run device server
func (s *server) Run() {

	s.logBuildInfo()

	if s.isSet(flagRunningSite) {
		s.logInfo("RUNNING full web site")
	} else if s.isSet(flagRunningDemo) {
		s.logInfo("RUNNING in DEMO mode")
	}

	// Install /model/{model} patterns for models
	s.installModels()

	if err := s.loadDevices(); err != nil {
		s.logError("Loading devices", "err", err)
		return
	}

	s.buildDevices()

	if err := s.buildTree(); err != nil {
		s.logError("Building tree", "err", err)
		return
	}

	s.root.set(flagOnline | flagMetal)

	if s.isSet(flagRunningDemo) {
		if err := s.root.demoSetup(); err != nil {
			s.logError("Setting up root device", "err", err)
			return
		}
	} else {
		if err := s.root.setup(); err != nil {
			s.logError("Setting up root device", "err", err)
			return
		}
	}

	// Dial parents
	s.dialParents(s.dialUrls, s.user, s.passwd)

	// If no port, don't run as a web server
	if s.port == 0 {
		s.root.run()
		s.logInfo("Bye, Bye", "root", s.root.Name)
		return
	}

	// Running as a web server...
	s.setupAPI()

	go s.gcViews()

	// Run http server in go routine to be shutdown later
	go func() {
		s.server.Addr = ":" + strconv.Itoa(s.port)
		s.logInfo("ListenAndServe", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logError("HTTP server ListenAndServe", "err", err)
			os.Exit(1)
		}
	}()

	// Ok, here we go...device should run until interrupted (or stopped)
	s.root.run()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.logError("HTTP server Shutdown", "err", err)
		os.Exit(1)
	}

	s.logInfo("Bye, Bye", "root", s.root.Name)
}

func (s *server) logBuildInfo() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		s.logDebug("Build Info:")
		s.logDebug("Go Version:", "version", buildInfo.GoVersion)
		s.logDebug("Path", "path", buildInfo.Path)
		for _, setting := range buildInfo.Settings {
			s.logDebug("Setting", setting.Key, setting.Value)
		}
		for _, dep := range buildInfo.Deps {
			s.logDebug("Dependency", "Path", dep.Path, "Version", dep.Version, "Replace", dep.Replace)
		}
	}
	s.logDebug("GOMAXPROCS", "n", runtime.GOMAXPROCS(0))
}
