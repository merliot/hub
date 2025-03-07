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

type server struct {
	devices         deviceMap      // key: device id, value: *device
	sessions        sessionMap     // key: session id, value: *session
	uplinks         uplinkMap      // key: linker
	downlinks       downlinkMap    // key: device id, value: linker
	models          ModelMap       // key: model name
	packetHandlers  PacketHandlers // key: path, value: handler
	root            *device
	mux             *http.ServeMux
	server          *http.Server
	saveToClipboard bool
	runningSite     bool // running as full web-site
	runningDemo     bool // running in DEMO mode
	wsxPingPeriod   int
}

var rlConfig = ratelimit.Config{
	RateLimitWindow: 100 * time.Millisecond,
	MaxRequests:     30,
	BurstSize:       30,
	CleanupInterval: 1 * time.Minute,
}

// NewServer returns a device server listening on addr
func NewServer(addr string, models ModelMap) *server {
	s := server{
		sessions:       newSessions(),
		models:         models,
		packetHandlers: make(PacketHandlers),
		mux:            http.NewServeMux(),
		server:         &http.Server{Addr: addr},
	}

	s.wsxPingPeriod, _ = strconv.Atoi(Getenv("PING_PERIOD", "2"))
	if s.wsxPingPeriod < 2 {
		s.wsxPingPeriod = 2
	}

	s.packetHandlers["/created"] = &PacketHandler[msgCreated]{s.handleCreated}
	s.packetHandlers["/destroyed"] = &PacketHandler[msgDestroy]{s.handleDestroyed}
	s.packetHandlers["/downloaded"] = &PacketHandler[msgDownloaded]{s.handleDownloaded}
	s.packetHandlers["/announced"] = &PacketHandler[deviceMap]{s.handleAnnounced}

	rl := ratelimit.New(rlConfig)
	s.server.Handler = rl.RateLimit(bassicAuth(s.mux))

	return &s
}

func (s *server) buildDevice(id string, d *device) error {
	if id != d.Id {
		return fmt.Errorf("Mismatching Ids")
	}
	if err := validateId(id); err != nil {
		return err
	}
	model, ok := s.models[d.Model]
	if !ok {
		return fmt.Errorf("Model '%s' not registered", d.Model)
	}
	return d.build(model.Maker, s.flags())
}

func (s *server) buildDevices() error {
	s.devices.drange(func(id string, d *device) bool {
		if err := s.buildDevice(id, d); err != nil {
			LogError("Skipping device", "id", id, "err", err)
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

	var err error

	logLevel = Getenv("LOG_LEVEL", "INFO")
	keepBuilds = Getenv("DEBUG_KEEP_BUILDS", "") == "true"
	s.runningSite = Getenv("SITE", "") == "true"
	runningDemo = (Getenv("DEMO", "") == "true") || s.runningSite

	logBuildInfo()

	if s.runningSite {
		LogInfo("RUNNING full web site")
	} else if s.runningDemo {
		LogInfo("RUNNING in DEMO mode")
	}

	if err := s.loadDevices(); err != nil {
		LogError("Loading devices", "err", err)
		return
	}

	if err := s.buildDevices(); err != nil {
		LogError("Building devices", "err", err)
		return
	}

	if err := s.buildTree(); err != nil {
		LogError("Building tree", "err", err)
		return
	}

	if err := s.root.setup(); err != nil {
		LogError("Setting up root device", "err", err)
		return
	}

	// Dial parents
	var urls = Getenv("DIAL_URLS", "")
	var user = Getenv("USER", "")
	var passwd = Getenv("PASSWD", "")
	dialParents(urls, user, passwd)

	// If Server.Addr empty, don't run as a web server
	if s.server.Addr == "" {
		s.root.run()
		LogInfo("Bye, Bye", "root", s.root.Name)
		return
	}

	// Running as a web server...
	s.setupAPI()

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
