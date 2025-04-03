// Merliot Hub
//
// To run standalone:
//   $ go run ./cmd
//
// To generate x84-64 and rpi binaries, run:
//   $ go generate ./cmd

//go:generate sh -x -c "go run ./gen-models/ ../models.json ../pkg/models/models.go"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ../bin/device-x86-64 -tags x86_64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOARM=5 go build -ldflags '-s -w' -o ../bin/device-rpi -tags rpi ./"

package main

import (
	"strconv"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

func main() {

	port, _ := strconv.Atoi(device.Getenv("PORT", "8000"))
	pingPeriod, _ := strconv.Atoi(device.Getenv("PING_PERIOD", "2"))
	keepBuilds := device.Getenv("DEBUG_KEEP_BUILDS", "")
	runningSite := device.Getenv("SITE", "")
	runningDemo := device.Getenv("DEMO", "")
	background := device.Getenv("BACKGROUND", "")
	wifiSsids := device.Getenv("WIFI_SSIDS", "")
	wifiPassphrases := device.Getenv("WIFI_PASSPHRASES", "")
	autoSave := device.Getenv("AUTO_SAVE", "true")
	devicesEnv := device.Getenv("DEVICES", "")
	devicesFile := device.Getenv("DEVICES_FILE", "")
	logLevel := device.Getenv("LOG_LEVEL", "INFO")
	dialUrls := device.Getenv("DIAL_URLS", "")
	user := device.Getenv("USER", "")
	passwd := device.Getenv("PASSWD", "")

	server := device.NewServer(
		device.WithPort(port),
		device.WithModels(models.AllModels),
		device.WithPingPeriod(pingPeriod),
		device.WithKeepBuilds(keepBuilds),
		device.WithRunningSite(runningSite),
		device.WithRunningDemo(runningDemo),
		device.WithBackground(background),
		device.WithWifiSsids(wifiSsids),
		device.WithWifiPassphrases(wifiPassphrases),
		device.WithAutoSave(autoSave),
		device.WithDevicesEnv(devicesEnv),
		device.WithDevicesFile(devicesFile),
		device.WithLogLevel(logLevel),
		device.WithDialUrls(dialUrls),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)

	server.Run()
}
