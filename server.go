//go:build !tinygo

package hub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var root *device

// Run the device
//
// Environment variables:
//
// PORT
// SITE
// DEMO
// WIFI_SSIDS
// WIFI_PASSPHRASES
// DEVICES
// DEVICES_FILE
// DEBUG_KEEP_BUILDS
// USER
// PASSWD
func Run() {

	var err error

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// TODO figure out why module version prints as (devel) and not v0.0.x
		fmt.Printf("Build Info: \n%s\n", buildInfo)
	}

	runningSite = (Getenv("SITE", "") == "true")
	runningDemo = (Getenv("DEMO", "") == "true") || runningSite

	if runningSite {
		fmt.Println("RUNNING full web site")
	} else if runningDemo {
		fmt.Println("RUNNING in DEMO mode")
	}

	if err := devicesLoad(); err != nil {
		fmt.Println("Error loading devices:", err)
		return
	}

	devicesBuild()

	root, err = devicesFindRoot()
	if err != nil {
		fmt.Println("Error finding root device:", err)
		return
	}

	if err := root.setup(); err != nil {
		fmt.Println("Error setting up root device:", err)
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
		log.Println("Device", root.Name, "done, bye")
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

	// Run http server in go routine to be shutdown later
	go func() {
		fmt.Println("ListenAndServe on", addr, "\n ")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}

	}()

	root.run()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	log.Println("Device", root.Name, "done, bye")
}
