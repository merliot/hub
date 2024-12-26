//go:build !tinygo

package device

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

var root *device

// Run the device
func Run() {

	var err error

	logBuildInfo()

	runningSite = (Getenv("SITE", "") == "true")
	runningDemo = (Getenv("DEMO", "") == "true") || runningSite

	if runningSite {
		LogInfo("RUNNING full web site")
	} else if runningDemo {
		LogInfo("RUNNING in DEMO mode")
	}

	if err := devicesLoad(); err != nil {
		LogError("Loading devices", "err", err)
		return
	}

	devicesBuild()

	root, err = findRoot(devices)
	if err != nil {
		LogError("Finding root device", "err", err)
		return
	}

	devicesSetupAPI()

	if err := root.setup(); err != nil {
		LogError("Setting up root device", "err", err)
		return
	}

	// Build route table from root's perpective
	routesBuild(root)

	// Dial parents
	var urls = Getenv("DIAL_URLS", "")
	var user = Getenv("USER", "")
	var passwd = Getenv("PASSWD", "")
	dialParents(urls, user, passwd)

	// If no port was given, don't run as a web server
	var port = Getenv("PORT", "8000")
	if port == "" {
		root.run()
		LogInfo("Bye, Bye", "root", root.Name)
		return
	}

	// Running as a web server...

	// Install /model/{model} patterns for makers
	modelsInstall()

	// Install the /device/{id} pattern for devices
	devicesInstall()

	// Install / to point to root device
	http.Handle("/", basicAuthHandler(root))

	// Install /ws websocket listener, but only if not in demo mode.  In
	// demo mode, we want to ignore any devices dialing in.
	if !runningDemo {
		http.HandleFunc("/ws", basicAuthHandlerFunc(wsHandle))
	}

	// Install /wsx websocket listener (wsx is for htmx ws)
	http.HandleFunc("/wsx", basicAuthHandlerFunc(wsxHandle))

	addr := ":" + port
	server := &http.Server{Addr: addr}

	http.HandleFunc("/devices", basicAuthHandlerFunc(showDevices))

	// Run http server in go routine to be shutdown later
	go func() {
		LogInfo("ListenAndServe", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			LogError("HTTP server ListenAndServe", "err", err)
			os.Exit(1)
		}

	}()

	// Ok, here we go...should run until interrupted
	root.run()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		LogError("HTTP server Shutdown", "err", err)
		os.Exit(1)
	}

	LogInfo("Bye, Bye", "root", root.Name)
}

func logBuildInfo() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		LogInfo("Build Info:")
		LogInfo("Go Version:", "version", buildInfo.GoVersion)
		LogInfo("Path", "path", buildInfo.Path)
		for _, setting := range buildInfo.Settings {
			LogInfo("Setting", setting.Key, setting.Value)
		}
		for _, dep := range buildInfo.Deps {
			LogInfo("Dependency", "Path", dep.Path, "Version", dep.Version, "Replace", dep.Replace)
		}
	}
	LogInfo("GOMAXPROCS", "n", runtime.GOMAXPROCS(0))
}
