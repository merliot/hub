//go:build !tinygo

package device

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/merliot/hub/pkg/ratelimit"
)

type server struct {
	devices         deviceMap // key: device id, value: *device
	root            *device
	models          ModelMap // key: model name
	mux             *http.ServeMux
	server          *http.Server
	saveToClipboard bool
}

var rlConfig = ratelimit.Config{
	RateLimitWindow: 100 * time.Millisecond,
	MaxRequests:     30,
	BurstSize:       30,
	CleanupInterval: 1 * time.Minute,
}

// NewServer returns a device server listening on addr
func NewServer(addr string) *server {
	s := server{
		models: make(ModelMap),
		mux:    http.NewServeMux(),
		server: &http.Server{Addr: addr},
	}
	rl := ratelimit.New(rlConfig)
	s.server.Handler = rl.RateLimit(bassicAuth(s.mux))
	return &s
}

func (s *server) buildDevice(id string, d *device) error {
	if id != d.Id {
		return fmt.Errorf("Mismatching Ids")
	}
	model, ok := s.models[d.Model]
	if !ok {
		return fmt.Errorf("Model '%s' not registered", d.Model)
	}
	return d.build(model.Maker)
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
	runningSite = Getenv("SITE", "") == "true"
	runningDemo = (Getenv("DEMO", "") == "true") || runningSite

	logBuildInfo()

	if runningSite {
		LogInfo("RUNNING full web site")
	} else if runningDemo {
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

	s.routesBuild()

	// Dial parents
	var urls = Getenv("DIAL_URLS", "")
	var user = Getenv("USER", "")
	var passwd = Getenv("PASSWD", "")
	dialParents(urls, user, passwd)

	// Default to port :8000
	var port = Getenv("PORT", "8000")

	// If port was explicitly not set, don't run as a web server
	if port == "" {
		s.root.run()
		LogInfo("Bye, Bye", "root", s.root.Name)
		return
	}

	// Running as a web server...

	// Install /model/{model} patterns for makers
	s.modelsInstall()

	// Install the /device/{id} pattern for devices
	s.devicesInstall()

	// Install / to point to root device
	s.mux.Handle("/", s.root)

	// Install /ws websocket listener, but only if not in demo mode.  In
	// demo mode, we want to ignore any devices dialing in.
	if !runningDemo {
		s.mux.HandleFunc("/ws", wsHandle)
	}

	// Install /wsx websocket listener (wsx is for htmx ws)
	s.mux.HandleFunc("/wsx", wsxHandle)

	s.mux.HandleFunc("/devices", showDevices)

	addr := ":" + port

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
