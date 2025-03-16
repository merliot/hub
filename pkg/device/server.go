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
	"time"

	"github.com/merliot/hub/pkg/ratelimit"
)

// Server flags
const (
	flagRunningDemo     flags = 1 << iota // Running in DEMO mode
	flagRunningSite                       // Running in SITE mode
	flagSaveToClipboard                   // Save changes to clipboard
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
	wsxPingPeriod int
}

var rlConfig = ratelimit.Config{
	FillInterval:    100 * time.Millisecond,
	Capacity:        30,
	CleanupInterval: 1 * time.Minute,
}

// NewServer returns a device server listening on addr
func NewServer(addr string, models Models) *server {
	s := server{
		sessions:       newSessions(),
		packetHandlers: make(PacketHandlers),
		mux:            http.NewServeMux(),
		server:         &http.Server{Addr: addr},
	}

	s.models.load(models)

	s.wsxPingPeriod, _ = strconv.Atoi(Getenv("PING_PERIOD", "2"))
	if s.wsxPingPeriod < 2 {
		s.wsxPingPeriod = 2
	}

	s.packetHandlers["/created"] = &PacketHandler[msgCreated]{s.handleCreated}
	s.packetHandlers["/destroyed"] = &PacketHandler[msgDestroy]{s.handleDestroyed}
	s.packetHandlers["/downloaded"] = &PacketHandler[msgDownloaded]{s.handleDownloaded}
	s.packetHandlers["/announced"] = &PacketHandler[deviceMap]{s.handleAnnounced}

	rl := ratelimit.New(rlConfig)
	s.server.Handler = rl.RateLimit(basicAuth(s.mux))

	if Getenv("DEBUG_KEEP_BUILDS", "") == "true" {
		s.set(flagDebugKeepBuilds)
	}

	if Getenv("SITE", "") == "true" {
		s.set(flagRunningSite)
	}

	if (Getenv("DEMO", "") == "true") || s.isSet(flagRunningSite) {
		s.set(flagRunningDemo)
	}

	return &s
}

func (s *server) newPacket() *Packet {
	return &Packet{server: s}
}

// routeDown routes the packet down to a downlink.  Which downlink is
// determined by a lookup in the routing table for the "next-hop" downlink, the
// downlink which is towards the destination.
func (s *server) routeDown(pkt *Packet) error {
	LogDebug("routeDown", "pkt", pkt)

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
	model, exists := s.models.get(d.Model)
	if !exists {
		return fmt.Errorf("Model '%s' not registered", d.Model)
	}
	d.model = model
	return d.build(s.defaultDeviceFlags())
}

func (s *server) buildDevices() {
	s.devices.drange(func(id string, d *device) bool {
		if err := s.buildDevice(id, d); err != nil {
			LogError("Skipping", "device", d, "err", err)
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

	logLevel = Getenv("LOG_LEVEL", "INFO")

	logBuildInfo()

	if s.isSet(flagRunningSite) {
		LogInfo("RUNNING full web site")
	} else if s.isSet(flagRunningDemo) {
		LogInfo("RUNNING in DEMO mode")
	}

	if err := s.loadDevices(); err != nil {
		LogError("Loading devices", "err", err)
		return
	}

	s.buildDevices()

	if err := s.buildTree(); err != nil {
		LogError("Building tree", "err", err)
		return
	}

	s.root.set(flagOnline | flagMetal)

	if s.isSet(flagRunningDemo) {
		if err := s.root.demoSetup(); err != nil {
			LogError("Setting up root device", "err", err)
			return
		}
	} else {
		if err := s.root.setup(); err != nil {
			LogError("Setting up root device", "err", err)
			return
		}
	}

	// Dial parents
	var urls = Getenv("DIAL_URLS", "")
	var user = Getenv("USER", "")
	var passwd = Getenv("PASSWD", "")
	s.dialParents(urls, user, passwd)

	// If Server.Addr empty, don't run as a web server
	if s.server.Addr == "" || s.server.Addr == ":" {
		s.root.run()
		LogInfo("Bye, Bye", "root", s.root.Name)
		return
	}

	// Running as a web server...
	s.setupAPI()

	go s.gcViews()

	// Run http server in go routine to be shutdown later
	go func() {
		LogInfo("ListenAndServe", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			LogError("HTTP server ListenAndServe", "err", err)
			os.Exit(1)
		}

	}()

	// Ok, here we go...should run until interrupted
	s.root.run()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		LogError("HTTP server Shutdown", "err", err)
		os.Exit(1)
	}

	LogInfo("Bye, Bye", "root", s.root.Name)
}

func logBuildInfo() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		LogDebug("Build Info:")
		LogDebug("Go Version:", "version", buildInfo.GoVersion)
		LogDebug("Path", "path", buildInfo.Path)
		for _, setting := range buildInfo.Settings {
			LogDebug("Setting", setting.Key, setting.Value)
		}
		for _, dep := range buildInfo.Deps {
			LogDebug("Dependency", "Path", dep.Path, "Version", dep.Version, "Replace", dep.Replace)
		}
	}
	LogDebug("GOMAXPROCS", "n", runtime.GOMAXPROCS(0))
}
