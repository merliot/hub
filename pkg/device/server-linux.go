//go:build !tinygo

package device

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/merliot/hub/pkg/ratelimit"
)

// Server
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
	serverFlags
	sync.Mutex
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

// ServerOption is a server option
type ServerOption func(*server)

// WithPort returns a ServerOption that sets the port number for the server to
// listen on.  The port must be a valid port number between 1024 and 49151
// (user port range).
func WithPort(port int) ServerOption {
	return func(s *server) {
		if port < 1024 || port > 49151 {
			port = 0 // invalid port, will run server without web server
		}
		s.port = port
	}
}

// WithModels returns a ServerOption that sets the device models available to
// the server.
func WithModels(models Models) ServerOption {
	return func(s *server) {
		s.models.load(models)
	}
}

// WithPingPeriod returns a ServerOption that sets the period (in seconds)
// between device pings.  This determines how often the server checks if
// devices are still connected.  Must be greater than or equal to 2s.  Default
// is 10s.
func WithPingPeriod(period int) ServerOption {
	return func(s *server) {
		if period < 2 {
			period = 2
		}
		s.wsxPingPeriod = period
	}
}

// WithKeepBuilds returns a ServerOption that sets whether to keep build
// artifacts after compilation.  The keep parameter should be "true" or "false"
// to enable or disable keeping builds.
func WithKeepBuilds(keep string) ServerOption {
	return func(s *server) {
		if keep == "true" {
			s.set(flagDebugKeepBuilds)
		}
	}
}

// WithRunningSite returns a ServerOption that enables full https://merliot.io
// website mode.
func WithRunningSite(running string) ServerOption {
	return func(s *server) {
		if running == "true" {
			s.set(flagRunningSite)
		}
	}
}

// WithRunningDemo returns a ServerOption that runs the server in demo mode.
// All devices will run in demo mode.
func WithRunningDemo(running string) ServerOption {
	return func(s *server) {
		if running == "true" {
			s.set(flagRunningDemo)
		}
	}
}

// WithBackground returns a ServerOption that sets whether the server should
// run in the background.  The bg parameter should be "GOOD" or "EVIL".  The
// default is "EVIL".
func WithBackground(bg string) ServerOption {
	return func(s *server) {
		s.background = bg
	}
}

// WithWifiSsids returns a ServerOption that sets the list of WiFi SSIDs to
// select from.  The ssids parameter should be a comma-separated list of SSID
// names.
func WithWifiSsids(ssids string) ServerOption {
	return func(s *server) {
		if ssids != "" {
			s.wifiSsids = strings.Split(ssids, ",")
		}
	}
}

// WithWifiPassphrases returns a ServerOption that sets the list of WiFi
// passphrases.  The ps parameter should be a comma-separated list of
// passphrases corresponding to the SSIDs.
func WithWifiPassphrases(ps string) ServerOption {
	return func(s *server) {
		if ps != "" {
			s.wifiPassphrases = strings.Split(ps, ",")
		}
	}
}

// WithAutoSave returns a ServerOption that sets whether device changes should
// be automatically saved.  The save parameter should be "true" or "false" to
// enable or disable auto-saving.
func WithAutoSave(save string) ServerOption {
	return func(s *server) {
		if save == "true" {
			s.set(flagAutoSave)
		}
	}
}

// WithDevicesEnv returns a ServerOption that sets the devices configuration
// from the DEVICES environment variable.  The devs parameter should be a JSON
// string containing the devices.
func WithDevicesEnv(devs string) ServerOption {
	return func(s *server) {
		s.devicesEnv = devs
	}
}

// WithDevicesFile returns a ServerOption that sets the path to the devices
// file.  The file parameter should be the path to a JSON file containing
// devices.
func WithDevicesFile(file string) ServerOption {
	return func(s *server) {
		s.devicesFile = file
	}
}

// WithLogLevel returns a ServerOption that sets the logging level for the
// server.  The level parameter should be one of: "DEBUG", "INFO", "WARN",
// "ERROR".  The default is "INFO".
func WithLogLevel(level string) ServerOption {
	return func(s *server) {
		s.logLevel = level
	}
}

// WithDialUrls returns a ServerOption that sets the list of URLs to dial.
// Each URL is a websocket URL in the form ws://host:port/ws or
// wss://host:port/ws.  The host:port is the address of the parent hub device.
// The URLs are used to dial parent devices.  Note that the server will dial
// all the URLs in the list, so one device can have multiple parents.  The urls
// parameter should be a comma-separated list of URLs.
func WithDialUrls(urls string) ServerOption {
	return func(s *server) {
		s.dialUrls = urls
	}
}

// WithUser returns a ServerOption that sets the username for HTTP Basic
// Authentication.
func WithUser(user string) ServerOption {
	return func(s *server) {
		s.user = user
	}
}

// WithPasswd returns a ServerOption that sets the password for HTTP Basic
// Authentication.
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

// NewServer returns new server.
func NewServer(options ...ServerOption) *server {

	s := &server{
		packetHandlers: make(PacketHandlers),
		mux:            http.NewServeMux(),
		server:         &http.Server{},
		wsxPingPeriod:  10,
		logLevel:       "INFO",
	}

	for _, opt := range options {
		opt(s)
	}

	if s.isSet(flagRunningSite) {
		s.set(flagRunningDemo)
	}

	s.packetHandlers["created"] = &PacketHandler[msgCreated]{s.handleCreated}
	s.packetHandlers["destroyed"] = &PacketHandler[msgDestroy]{s.handleDestroyed}
	s.packetHandlers["downloaded"] = &PacketHandler[msgDownloaded]{s.handleDownloaded}
	s.packetHandlers["announced"] = &PacketHandler[deviceMap]{s.handleAnnounced}

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

func (s *server) newDevice(id, model, name, params string) (d *device, err error) {
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
		Id:           id,
		Model:        model,
		Name:         name,
		DeployParams: params,
		model:        m,
		server:       s,
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

type serverStatus struct {
	Flags     string
	Sessions  sessionsJSON
	Uplinks   linksJSON
	Downlinks downlinksJSON
}

func (s *server) statusJSON() []byte {
	var status = serverStatus{
		Flags:     s.serverFlags.list(),
		Sessions:  s.sessions.getJSON(),
		Uplinks:   s.uplinks.getJSON(),
		Downlinks: s.downlinks.getJSON(),
	}
	j, _ := json.MarshalIndent(&status, "", "\t")
	return j
}
